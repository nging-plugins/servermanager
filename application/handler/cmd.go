/*
   Nging is a toolbox for webmasters
   Copyright (C) 2018-present  Wenhui Shen <swh@admpub.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package handler

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/admpub/sockjs-go/v3/sockjs"
	"github.com/admpub/websocket"
	"github.com/webx-top/com"
	"github.com/webx-top/db"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/defaults"
	"github.com/webx-top/echo/param"

	"github.com/admpub/gopty"
	"github.com/admpub/log"
	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/charset"
	"github.com/coscms/webcore/library/common"
	"github.com/coscms/webcore/library/config"
	"github.com/coscms/webcore/library/cron"

	"github.com/nging-plugins/servermanager/application/dbschema"
	conf "github.com/nging-plugins/servermanager/application/library/config"
	"github.com/nging-plugins/servermanager/application/model"
	sshdbschema "github.com/nging-plugins/sshmanager/application/dbschema"
	sshmodel "github.com/nging-plugins/sshmanager/application/model"
)

func Cmd(ctx echo.Context) error {
	var err error
	id := ctx.Formx(`id`).Uint()
	m := model.NewCommand(ctx)
	if id > 0 {
		err := m.Get(nil, `id`, id)
		if err != nil {
			common.SendFail(ctx, err.Error())
			return ctx.Redirect(backend.URLFor(`/manager/command`))
		}
	}
	ctx.Set(`id`, id)
	ctx.Set(`cmd`, m.Command)
	ctx.Set(`isWindows`, com.IsWindows)
	ctx.Set(`isLinux`, com.IsLinux)
	ctx.Set(`isMac`, com.IsMac)
	return ctx.Render(`server/cmd`, err)
}

func CmdSendBySockJS(c sockjs.Session) error {
	send := make(chan string)
	//push(writer)
	go func() {
		for {
			message := <-send
			backend.WebSocketLogger.Debug(`Push message: `, message)
			if err := c.Send(message); err != nil {
				backend.WebSocketLogger.Error(`Push error: `, err.Error())
				return
			}
		}
	}()
	timeout := c.Request().URL.Query().Get(`timeout`)
	duration := config.ParseTimeDuration(timeout)
	//echo
	exec := func(session sockjs.Session) error {
		var (
			w   io.WriteCloser
			cmd *exec.Cmd
		)
		for {
			command, err := session.Recv()
			if err != nil {
				return err
			}
			if len(command) == 0 {
				continue
			}
			var (
				workdir   string
				env       []string
				hasRemote bool
			)
			if command[0] == '>' {
				id := param.String(command[1:]).Uint()
				if id > 0 {
					m, result, err := ExecCommand(id)
					if err != nil {
						send <- remoteCmdResultPrefix() + err.Error()
						continue
					}
					if m.Remote == `Y` { // 仅在远程执行的命令
						send <- remoteCmdResultPrefix() + result
						hasRemote = true
						continue
					}
					if m.Remote == `A` { // 同时在远程和本地执行的命令
						send <- remoteCmdResultPrefix() + result
						hasRemote = true
					}
					workdir = m.WorkDirectory
					env = conf.ParseEnvSlice(m.Env)
					command = m.Command
				} else {
					return errors.New(`Invalid ID: ` + command[1:])
				}
			} else {

			}
			if w == nil {
				w, cmd, err = cmdRunner(c.Request().Context(), workdir, env, command, send, func() {
					w.Close()
					w = nil
				}, duration, hasRemote)
				if err != nil {
					return err
				}
				continue
			}
			err = cmdContinue(command, w, cmd)
			if err != nil {
				return err
			}
		}
	}
	err := exec(c)
	if err != nil {
		backend.WebSocketLogger.Error(err)
	}
	close(send)
	return nil
}

func cmdRunner(ctx context.Context, workdir string, env []string, command string, send chan string, onEnd func(), timeout time.Duration, hasRemote bool) (w io.WriteCloser, cmd *exec.Cmd, err error) {
	params := cron.CmdParams(command)
	var prefixSended atomic.Bool
	output := func(b []byte) (e error) {
		defer func() {
			if re := recover(); re != nil {
				if cmd == nil {
					return
				}
				if cmd.Process == nil {
					return
				}
				err := cmd.Process.Kill()
				if err != nil {
					log.Warn(err)
				}
			}
		}()
		if com.IsWindows {
			b, e = charset.Convert(`gbk`, `utf-8`, b)
			if e != nil {
				return e
			}
		}
		if hasRemote && !prefixSended.Swap(true) {
			send <- localCmdResultPrefix() + string(b)
			return nil
		}
		send <- string(b)
		return nil
	}
	cmd = com.CreateCmd(params, output)
	cmd.Stderr = com.CmdResultCapturer{Do: output}
	if len(workdir) > 0 {
		cmd.Dir = workdir
	}
	if len(env) > 0 {
		cmd.Env = env
	}
	w, err = cmd.StdinPipe()
	if err != nil {
		return
	}
	done := make(chan struct{})
	go func() {
		log.Info(`[command] running: `, command)
		if e := cmd.Run(); e != nil {
			cmd.Stderr.Write([]byte(e.Error()))
		}
		done <- struct{}{}
		onEnd()
	}()
	cmdMaxTimeout := config.FromFile().Sys.CmdTimeoutDuration
	if timeout <= 0 {
		timeout = time.Minute * 5
	}
	if timeout > cmdMaxTimeout {
		timeout = cmdMaxTimeout
	}
	go func() {
		ticker := time.NewTicker(timeout)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				log.Info(`[command] exited: `, command)
				return
			case <-ticker.C:
				if cmd == nil {
					return
				}
				if cmd.Process == nil {
					return
				}
				cmd.Stderr.Write([]byte(`timeout`))
				log.Info(`[command] timeout: `, command)
				err := cmd.Process.Kill()
				if err != nil {
					log.Warn(err)
				}
				return
			case <-ctx.Done():
				if cmd == nil {
					return
				}
				if cmd.Process == nil {
					return
				}
				cmd.Stderr.Write([]byte(`request is cancelled`))
				log.Info(`[command] request is cancelled: `, command)
				err := cmd.Process.Kill()
				if err != nil {
					log.Warn(err)
				}
				return
			}
		}
	}()
	return
}

func cmdContinue(command string, w io.WriteCloser, cmd *exec.Cmd) (err error) {
	if cmd == nil {
		return nil
	}
	switch command {
	case `^C`:
		err = cmd.Process.Signal(os.Interrupt)
		if err != nil {
			if !strings.HasPrefix(err.Error(), `not supported by `) {
				backend.WebSocketLogger.Error(err)
			}
			err = cmd.Process.Kill()
			if err != nil {
				backend.WebSocketLogger.Error(err)
			}
		}
	default:
		w.Write([]byte(command + "\n"))
	}
	return nil
}

func Pty(c *websocket.Conn, ctx echo.Context) error {
	return gopty.ServeWebsocket(c, 120, 60)
}

func CmdSendByWebsocket(c *websocket.Conn, ctx echo.Context) error {
	check, _ := ctx.Funcs()[`CheckPerm`].(func(string) error)
	send := make(chan string)
	//push(writer)
	go func() {
		for {
			message := <-send
			backend.WebSocketLogger.Debug(`Push message: `, message)
			if err := c.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				backend.WebSocketLogger.Error(`Push error: `, err.Error())
				return
			}
		}
	}()

	timeout := ctx.Query(`timeout`)
	duration := config.ParseTimeDuration(timeout)
	//echo
	exec := func(conn *websocket.Conn) error {
		var (
			w   io.WriteCloser
			cmd *exec.Cmd
		)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}
			command := string(message)
			if len(command) == 0 {
				continue
			}
			//TEST:
			//notice.OpenMessage(`test`, ``)
			//notice.Send(`test`, notice.NewMessageWithValue(``, `from: admin`, `test user message`))
			var (
				workdir   string
				env       []string
				hasRemote bool
			)
			if command[0] == '>' {
				id := param.String(command[1:]).Uint()
				if id > 0 {
					if check != nil {
						err := check(command[1:])
						if err != nil {
							send <- err.Error()
							continue
						}
					}
					m, result, err := ExecCommand(id)
					if err != nil {
						send <- remoteCmdResultPrefix() + err.Error()
						continue
					}
					if m.Remote == `Y` {
						send <- remoteCmdResultPrefix() + result
						hasRemote = true
						continue
					}
					if m.Remote == `A` {
						send <- remoteCmdResultPrefix() + result
						hasRemote = true
					}
					workdir = m.WorkDirectory
					env = conf.ParseEnvSlice(m.Env)
					command = m.Command
				} else {
					err := errors.New(`Invalid ID: ` + command[1:])
					send <- err.Error()
					continue
				}
			} else {
				if check != nil {
					err := check(``)
					if err != nil {
						return err
					}
				}
			}
			if w == nil {
				w, cmd, err = cmdRunner(ctx.Request().StdRequest().Context(), workdir, env, command, send, func() {
					w.Close()
					w = nil
				}, duration, hasRemote)
				if err != nil {
					return err
				}
				continue
			}
			err = cmdContinue(command, w, cmd)
			if err != nil {
				return err
			}
		}
	}
	err := exec(c)
	if err != nil {
		backend.WebSocketLogger.Error(err)
	}
	close(send)
	return nil
}

func ExecCommand(id uint) (*dbschema.NgingCommand, string, error) {
	ctx := defaults.NewMockContext()
	m := model.NewCommand(ctx)
	err := m.Get(nil, `id`, id)
	if err != nil {
		return m.NgingCommand, "", err
	}
	if m.NgingCommand.Disabled == `Y` {
		return m.NgingCommand, "", errors.New(echo.T(`该命令已禁用`))
	}
	//m.NgingCommand.Remote = `Y`
	//m.NgingCommand.SshAccountId = 4
	if m.NgingCommand.Remote == `Y` || m.NgingCommand.Remote == `A` {
		if len(m.NgingCommand.SshAccountId) < 1 {
			return m.NgingCommand, "", errors.New("Error, you did not choose ssh account")
		}
		accountIDs := param.Split(m.NgingCommand.SshAccountId, `,`).Uint(param.IsGreaterThanZeroElement[uint])
		var cmdList []string
		if len(m.WorkDirectory) > 0 {
			cmdList = append(cmdList, `cd `+m.WorkDirectory)
		}
		if len(m.Env) > 0 {
			for _, env := range strings.Split(m.Env, "\n") {
				env = strings.TrimSpace(env)
				if len(env) < 1 {
					continue
				}
				cmdList = append(cmdList, `export `+env)
			}
		}
		cmdList = append(cmdList, m.NgingCommand.Command)
		sshUser := sshmodel.NewSshUser(ctx)
		_, err = sshUser.ListByOffset(nil, nil, 0, -1, `id`, db.In(accountIDs))
		if err != nil {
			if err == db.ErrNoMoreRows {
				return m.NgingCommand, "", errors.New("The specified ssh account does not exist")
			}
			return m.NgingCommand, "", err
		}
		rows := sshUser.Objects()
		if len(rows) == 0 {
			return m.NgingCommand, "", errors.New("The specified ssh account does not exist")
		}
		results := make([]string, len(rows))
		w := cron.NewCmdRec(1000)
		for idx, row := range rows {
			row.SetContext(ctx)
			resultTitle := `###### ` + row.Name + ` (` + row.Host + `) #####`
			err := sshExecCmd(row, cmdList, w)
			if err != nil {
				results[idx] = resultTitle + com.StrLF + err.Error()
			} else {
				results[idx] = resultTitle + com.StrLF + w.String()
			}
			w.Reset()
		}
		return m.NgingCommand, strings.Join(results, com.StrLF+com.StrLF), nil
	}
	return m.NgingCommand, "", err
}

func sshExecCmd(sshUser *sshdbschema.NgingSshUser, cmdList []string, w io.Writer) error {
	return sshmodel.ExecMultiCMD(sshUser, w, cmdList...)
}

func remoteCmdResultPrefix() string {
	out := "================================================\n"
	out += "==================[REMOTE]======================\n"
	out += "================================================\n"
	return out
}

func localCmdResultPrefix() string {
	out := "================================================\n"
	out += "===================[LOCAL]======================\n"
	out += "================================================\n"
	return out
}
