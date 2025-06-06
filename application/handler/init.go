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
	"github.com/webx-top/echo"
	ws "github.com/webx-top/echo/handler/websocket"

	"github.com/coscms/webcore/library/config"
	"github.com/coscms/webcore/library/httpserver"
	"github.com/coscms/webcore/library/module"
)

func RegisterRoute(r module.Router) {
	r.Backend().RegisterToGroup(`/server`, registerRoute)
}

func registerRoute(g echo.RouteRegister) {
	g.Route("GET", `/sysinfo`, Info)
	g.Route("GET", `/netstat`, Connections)
	g.Route("GET", `/processes`, ProcessList)
	g.Route("GET", `/process/:pid`, ProcessInfo)
	g.Route("GET", `/procskill/:pid`, ProcessKill)
	g.Route(`GET,POST`, `/service`, Service)
	g.Route(`GET,POST`, `/hosts`, Hosts)
	g.Route(`GET,POST`, `/daemon_index`, DaemonIndex)
	g.Route(`GET,POST`, `/daemon_add`, DaemonAdd)
	g.Route(`GET,POST`, `/daemon_edit`, DaemonEdit)
	g.Route(`GET,POST`, `/daemon_delete`, DaemonDelete)
	g.Route(`GET,POST`, `/daemon_restart`, DaemonRestart)
	g.Route("GET", `/cmd`, Cmd)
	g.Route(`GET,POST`, `/command`, Command)
	g.Route(`GET,POST`, `/command_add`, CommandAdd)
	g.Route(`GET,POST`, `/command_edit`, CommandEdit)
	g.Route(`GET,POST`, `/command_delete`, CommandDelete)
	g.Route(`GET,POST`, `/daemon_log`, DaemonLog)
	g.Route(`GET,POST`, `/log/:category`, func(c echo.Context) error {
		return config.FromFile().Settings().Log.Show(c)
	})
	g.Get(`/status`, Status).SetMetaKV(httpserver.PermPublicKV())
	//sockjsHandler.New("/cmdSend",CmdSendBySockJS).Wrapper(g)
	ws.New("/cmdSendWS", CmdSendByWebsocket).Wrapper(g)
	ws.New("/ptyWS", Pty).Wrapper(g)
	ws.New("/dynamic", InfoByWebsocket).Wrapper(g).SetMetaKV(httpserver.PermPublicKV())
}
