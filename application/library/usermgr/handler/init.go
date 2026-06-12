//go:build linux

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

	"github.com/nging-plugins/servermanager/application/handler"
)

func init() {
	handler.LeftNavigate.Children.Add(-1, usermgrLeftNavigate...)
	handler.AddRouteRegister(registerUserRoute)
}

func registerUserRoute(g echo.RouteRegister) {
	g.Route(`GET`, `/system_user`, SystemUserList)
	g.Route(`GET,POST`, `/system_user/add`, SystemUserAdd)
	g.Route(`GET,POST`, `/system_user/edit`, SystemUserEdit)
	g.Route(`POST`, `/system_user/delete`, SystemUserDelete)
	g.Route(`POST`, `/system_user/lock`, SystemUserLock)
	g.Route(`POST`, `/system_user/unlock`, SystemUserUnlock)
}
