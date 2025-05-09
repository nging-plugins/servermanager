package handler

import (
	"context"
	"testing"

	"github.com/admpub/pp"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/stretchr/testify/assert"
)

func TestNetstat(t *testing.T) {
	conns, err := net.ConnectionsWithContext(context.Background(), `all`)
	assert.NoError(t, err)
	pp.Println(conns)
}
