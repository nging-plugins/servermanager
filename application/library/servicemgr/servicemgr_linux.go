package servicemgr

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/coreos/go-systemd/dbus"
	dbusLib "github.com/godbus/dbus"
)

const serviceSuffix = ".service"

func NewClient() (*Client, error) {
	conn, err := dbus.New()
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

// Client represents systemd D-Bus API client.
type Client struct {
	conn    *dbus.Conn
	runtime bool
}

func (c *Client) List() ([]*Service, error) {
	var list []*Service
	units, err := c.conn.ListUnitsByPatterns(nil, []string{`*.service`})
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

func (c *Client) listFiles(names []string) ([]dbus.UnitFile, error) {
	return c.conn.ListUnitFilesByPatterns(nil, names)
}

func (c *Client) SetRuntime(runtime bool) {
	c.runtime = runtime
}

func (c *Client) getServiceName(name string) string {
	return getServiceName(name)
}

func (c *Client) getFilesByName(name string) ([]string, error) {
	name = c.getServiceName(name)
	unitFiles, err := c.listFiles([]string{name})
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

func (c *Client) Enable(name string) error {
	files, err := c.getFilesByName(name)
	if err != nil {
		return err
	}
	_, _, err = c.conn.EnableUnitFiles(files, c.runtime, false)
	return err
}

func (c *Client) Disable(name string) error {
	files, err := c.getFilesByName(name)
	if err != nil {
		return err
	}
	_, err = c.conn.DisableUnitFiles(files, c.runtime)
	return err
}

func (c *Client) Start(name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.StartUnit(name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) Stop(name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.StopUnit(name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) Restart(name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.RestartUnit(name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) ReloadOrRestart(name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.ReloadOrRestartUnit(name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) ReloadUnit(name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.ReloadUnit(name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (c *Client) StopAndStart(name string) error {
	ch := make(chan string)
	name = c.getServiceName(name)
	_, err := c.conn.StopUnit(name, `replace`, ch)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

// Reload instructs systemd to scan for and reload unit files. This is
// equivalent to a 'systemctl daemon-reload'.
func (c *Client) Reload() error {
	err := c.conn.Reload()
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

func getServiceName(name string) string {
	if !strings.HasSuffix(name, serviceSuffix) {
		name += serviceSuffix
	}
	return name
}

func ServiceLog(ctx context.Context, service string, callback func(rd io.Reader) error, follow ...bool) error {
	args := []string{
		`-u`, getServiceName(service),
		`-n`, `100`,
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
