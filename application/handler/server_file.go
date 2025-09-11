//go:build !windows

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
	"strings"

	"github.com/webx-top/echo"

	"github.com/coscms/webcore/library/filemanager"
	"github.com/coscms/webcore/library/filemanager/filemanagerhandler"
	"github.com/coscms/webcore/library/navigate"

	routeRegistry "github.com/coscms/webcore/registry/route"
)

func init() {
	LeftNavigate.Children.Add(-1, &navigate.Item{
		Display: true,
		Name:    echo.T(`服务器文件管理`),
		Action:  `file`,
		Icon:    `file`,
	})
}

func registerRouteServerFile(g echo.RouteRegister) {
	metaHandler := routeRegistry.IRegister().MetaHandler
	g.Route(`GET,POST`, `/file`, metaHandler(echo.H{`name`: `管理服务器文件`}, ServerFile))
}

func ServerFile(ctx echo.Context) error {
	root := `/`
	urlPrefix := ctx.Request().URL().Path() + `?path=` + filemanager.EncodedSepa
	h := filemanagerhandler.New(root, urlPrefix)
	h.SetCanChmod(true)
	h.SetCanChown(true)
	err := h.Handle(ctx)
	if err != nil || ctx.Response().Committed() {
		return err
	}
	rootPath := strings.TrimSuffix(root, echo.FilePathSeparator)
	if len(rootPath) == 0 {
		rootPath = root
		ctx.Set(`rootPath`, rootPath)
	}
	ctx.SetFunc(`PermInfo`, filemanager.FileModeToPerms)
	ctx.Set(`activeURL`, `/server/file`)
	return ctx.Render(`server/file`, err)
}
