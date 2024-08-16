package servicemgr

import (
	"bufio"
	"context"
	"io"
	"os/exec"
)

func List(ctx context.Context) (r []*Service, e error) {
	cmd := exec.CommandContext(
		ctx,
		`systemctl`, `list-units`, `--type`, `service`, // linux
		//`launchctl`, `list`, // macOS
	)
	e = ReadCmdOutput(cmd, func(rd io.ReadCloser) error {
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

func ServiceLog(ctx context.Context, service string, callback func(rd io.ReadCloser) error, follow ...bool) error {
	args := []string{
		`-u`, service,
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
