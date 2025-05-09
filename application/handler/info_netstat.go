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
	"github.com/coscms/webcore/library/common"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/webx-top/echo"
)

func Connections(ctx echo.Context) (err error) {
	var conns []net.ConnectionStat
	kind := ctx.Form(`kind`, `all`)
	switch kind {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "unix", "inet", "inet4", "inet6":
	default:
		kind = "all"
	}
	conns, err = net.ConnectionsWithContext(ctx, kind)
	ctx.Set(`listData`, conns)
	return ctx.Render(`server/netstat`, common.Err(ctx, err))
}
