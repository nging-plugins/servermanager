package servicemgr

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webx-top/com"
)

func TestParse(t *testing.T) {
	output := `UNIT                               LOAD   ACTIVE SUB     DESCRIPTION
apparmor.service                   loaded active exited  Load AppArmor profiles
chrony.service                     loaded active running chrony, an NTP client/server
cloud-config.service               loaded active exited  Apply the settings specified in cloud-config
cloud-final.service                loaded active exited  Execute cloud user/final scripts
cloud-init-local.service           loaded active exited  Initial cloud-init job (pre-networking)
cloud-init.service                 loaded active exited  Initial cloud-init job (metadata service crawler)
console-setup.service              loaded active exited  Set console font and keymap`
	lines := strings.Split(output, "\n")
	r := []*Service{}
	for _, line := range lines {
		s := Parse(line)
		if s == nil {
			continue
		}
		r = append(r, s)
	}
	com.Dump(r)
}

// *
func TestList(t *testing.T) {
	ctx := context.Background()
	list, err := List(ctx)
	assert.NoError(t, err)
	com.Dump(list)
	client, err := NewClient(ctx)
	assert.NoError(t, err)
	defer client.Close()
	list, err = client.List(ctx, []string{}, []string{})
	assert.NoError(t, err)
	com.Dump(list)
	// err = client.Enable(ctx, "Nging")
	// assert.NoError(t, err)
	ServiceLog(ctx, "ssh", func(rd io.Reader) error {
		buf := make([]byte, 4096)
		n, _ := rd.Read(buf)
		println(string(buf[:n]))
		return nil
	}, false)
}

//*/
