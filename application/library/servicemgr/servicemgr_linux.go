package servicemgr

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	dbusLib "github.com/godbus/dbus/v5"
	"github.com/webx-top/com"
)

const serviceSuffix = ".service"

func NewClient(ctx context.Context) (*Client, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, runtime: true}, nil
}

// Client represents systemd D-Bus API client.
type Client struct {
	conn    *dbus.Conn
	runtime bool
}

func (c *Client) List(ctx context.Context, states []string, patterns []string) ([]*Service, error) {
	var list []*Service
	if len(patterns) == 0 {
		patterns = []string{`*.service`}
	}
	units, err := c.conn.ListUnitsByPatternsContext(ctx, states, patterns)
	if err != nil {
		return list, err
	}
	for _, unit := range units {
		s := &Service{}
		s.Name = strings.TrimSuffix(unit.Name, serviceSuffix)
		s.Load = unit.LoadState
		s.Active = unit.ActiveState
		s.Sub = unit.SubState
		s.Description = unit.Description
		list = append(list, s)
	}
	return list, nil
}

func (c *Client) ListFiles(ctx context.Context, names []string) ([]dbus.UnitFile, error) {
	return c.conn.ListUnitFilesByPatternsContext(ctx, nil, names)
}

func (c *Client) SetRuntime(runtime bool) {
	c.runtime = runtime
}

func (c *Client) getServiceName(name string) string {
	return GetServiceName(name)
}

func (c *Client) getFilesByName(ctx context.Context, name string) ([]string, error) {
	name = c.getServiceName(name)
	unitFiles, err := c.ListFiles(ctx, []string{name})
	if err != nil {
		return nil, err
	}
	if len(unitFiles) == 0 {
		return nil, dbusLib.ErrMsgNoObject
	}
	files := make([]string, 0, len(unitFiles))
	for _, uf := range unitFiles {
		files = append(files, uf.Path)
	}
	return files, err
}

func (c *Client) Enable(ctx context.Context, name string) error {
	files, err := c.getFilesByName(ctx, name)
	if err != nil {
		return err
	}
	_, _, err = c.conn.EnableUnitFilesContext(ctx, files, c.runtime, false)
	return err
}

func (c *Client) Disable(ctx context.Context, name string) error {
	files, err := c.getFilesByName(ctx, name)
	if err != nil {
		return err
	}
	_, err = c.conn.DisableUnitFilesContext(ctx, files, c.runtime)
	return err
}

func (c *Client) Start(ctx context.Context, name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.StartUnitContext(ctx, name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) Stop(ctx context.Context, name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.StopUnitContext(ctx, name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) Restart(ctx context.Context, name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.RestartUnitContext(ctx, name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) ReloadOrRestart(ctx context.Context, name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.ReloadOrRestartUnitContext(ctx, name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) ReloadUnit(ctx context.Context, name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.ReloadUnitContext(ctx, name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

// Reload instructs systemd to scan for and reload unit files. This is
// equivalent to a 'systemctl daemon-reload'.
func (c *Client) Reload(ctx context.Context) error {
	err := c.conn.ReloadContext(ctx)
	return err
}

func (c *Client) Close() error {
	c.conn.Close()
	return nil
}

func List(ctx context.Context) (r []*Service, e error) {
	cmd := exec.CommandContext(
		ctx,
		`systemctl`, `list-units`, `--type`, `service`, // linux
		//`launchctl`, `list`, // macOS
	)
	e = ReadCmdOutput(cmd, func(rd io.Reader) error {
		sc := bufio.NewScanner(rd)
		for sc.Scan() {
			line := sc.Text()
			//println(line)
			s := Parse(line)
			if s != nil {
				r = append(r, s)
			}
		}
		return nil
	})
	return
}

func GetServiceName(name string) string {
	if !strings.HasSuffix(name, serviceSuffix) {
		name += serviceSuffix
	}
	return name
}

func ServiceLog(ctx context.Context, service string, callback func(rd io.Reader) error, follow ...bool) error {
	return ServiceLogWithRows(ctx, service, 100, callback, follow...)
}

func ServiceLogWithRows(ctx context.Context, service string, lines uint, callback func(rd io.Reader) error, follow ...bool) error {
	args := []string{
		`-u`, GetServiceName(service),
		`-n`, com.String(lines),
	}
	if len(follow) > 0 && follow[0] {
		args = append(args, `-f`)
	}
	cmd := exec.CommandContext(
		ctx, `journalctl`,
		args...,
	)
	return ReadCmdOutput(cmd, callback)
}
