package servicemgr

import (
	"io"
	"os/exec"
	"strings"

	"github.com/webx-top/com"
)

type Service struct {
	Name        string
	Load        string
	Active      string
	Sub         string
	Type        string
	Description string
}

func Parse(line string) *Service {
	// UNIT                              LOAD  ACTIVE SUB    DESCRIPTION
	fields := com.ParseFields(line)
	if len(fields) > 0 && strings.HasSuffix(fields[0], `.service`) {
		fields[0] = strings.TrimSuffix(fields[0], `.service`)
		s := &Service{}
		com.SliceExtract(fields, &s.Name, &s.Load, &s.Active, &s.Sub)
		if len(fields) > 4 {
			s.Description = strings.Join(fields[4:], ` `)
		}
		return s
	}
	return nil
}

func ReadCmdOutput(cmd *exec.Cmd, callback func(rd io.Reader) error) error {
	rd, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer rd.Close()
	done := make(chan struct{})
	go func() {
		err = cmd.Run()
		done <- struct{}{}
		close(done)
	}()
	if _err := callback(rd); _err != nil {
		err = _err
	}
	<-done
	return err
}
