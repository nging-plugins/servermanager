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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	uploadClient "github.com/webx-top/client/upload"
	uploadDropzone "github.com/webx-top/client/upload/driver/dropzone"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"
	"github.com/coscms/webcore/library/config"
	"github.com/coscms/webcore/library/filemanager"
	"github.com/coscms/webcore/library/navigate"
	"github.com/coscms/webcore/library/notice"
	"github.com/coscms/webcore/library/respond"
	uploadChunk "github.com/coscms/webcore/registry/upload/chunk"

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
	var err error
	filePath := ctx.Form(`path`)
	do := ctx.Form(`do`)
	root := `/`
	mgr := filemanager.New(root, config.FromFile().Sys.EditableFileMaxBytes(), ctx)
	absPath := root
	user := backend.User(ctx)
	if len(filePath) > 0 {
		filePath = filepath.Clean(filePath)
		absPath = filepath.Join(root, filePath)
	}

	switch do {
	case `edit`:
		data := ctx.Data()
		if _, ok := Editable(absPath); !ok {
			data.SetInfo(ctx.T(`此文件不能在线编辑`), 0)
		} else {
			content := ctx.Form(`content`)
			encoding := ctx.Form(`encoding`)
			dat, err := mgr.Edit(absPath, content, encoding)
			if err != nil {
				data.SetInfo(err.Error(), 0)
			} else {
				data.SetData(dat, 1)
			}
		}
		return ctx.JSON(data)
	case `rename`:
		data := ctx.Data()
		newName := ctx.Form(`name`)
		err = mgr.Rename(absPath, newName)
		if err != nil {
			data.SetInfo(err.Error(), 0)
		} else {
			data.SetCode(1)
		}
		return ctx.JSON(data)
	case `mkdir`:
		data := ctx.Data()
		newName := ctx.Form(`name`)
		err = mgr.Mkdir(filepath.Join(absPath, newName), os.ModePerm)
		if err != nil {
			data.SetInfo(err.Error(), 0)
		} else {
			data.SetCode(1)
		}
		return ctx.JSON(data)
	case `delete`:
		paths := ctx.FormValues(`path`)
		next := ctx.Referer()
		if len(next) == 0 {
			next = ctx.Request().URL().Path() + fmt.Sprintf(`?path=%s`, com.URLEncode(filepath.Dir(filePath)))
		}
		for _, filePath := range paths {
			filePath = strings.TrimSpace(filePath)
			if len(filePath) == 0 {
				continue
			}
			filePath = filepath.Clean(filePath)
			absPath = filepath.Join(root, filePath)
			err = mgr.Remove(absPath)
			if err != nil {
				common.SendFail(ctx, err.Error())
				return ctx.Redirect(next)
			}
		}
		return ctx.Redirect(next)
	case `upload`:
		var cu *uploadClient.ChunkUpload
		var opts []uploadClient.ChunkInfoOpter
		if user != nil {
			cu = uploadChunk.NewUploader(fmt.Sprintf(`user/%d`, user.Id))
			opts = append(opts, uploadClient.OptChunkInfoMapping(uploadDropzone.MappingChunkInfo))
		}
		err = mgr.Upload(absPath, cu, opts...)
		if err != nil {
			user := backend.User(ctx)
			if user != nil {
				notice.OpenMessage(user.Username, `upload`)
				notice.Send(user.Username, notice.NewMessageWithValue(`upload`, ctx.T(`文件上传出错`), err.Error()))
			}
		}
		return respond.Dropzone(ctx, err, nil)
	default:
		var dirs []os.FileInfo
		var exit bool
		err, exit, dirs = mgr.List(absPath)
		if exit {
			return err
		}
		ctx.Set(`dirs`, dirs)
	}
	if filePath == `.` {
		filePath = ``
	}
	pathSlice := strings.Split(strings.Trim(filePath, echo.FilePathSeparator), echo.FilePathSeparator)
	pathLinks := make(echo.KVList, len(pathSlice))
	encodedSep := filemanager.EncodedSepa
	urlPrefix := ctx.Request().URL().Path() + `?path=` + encodedSep
	for k, v := range pathSlice {
		urlPrefix += com.URLEncode(v)
		pathLinks[k] = &echo.KV{K: v, V: urlPrefix}
		urlPrefix += encodedSep
	}
	ctx.Set(`pathLinks`, pathLinks)
	rootPath := strings.TrimSuffix(root, echo.FilePathSeparator)
	if len(rootPath) == 0 {
		rootPath = root
	}
	ctx.Set(`rootPath`, rootPath)
	ctx.Set(`path`, filePath)
	ctx.Set(`absPath`, absPath)
	ctx.SetFunc(`Editable`, func(fileName string) bool {
		_, ok := Editable(fileName)
		return ok
	})
	ctx.SetFunc(`Playable`, func(fileName string) string {
		mime, _ := Playable(fileName)
		return mime
	})
	ctx.Set(`activeURL`, `/server/file`)
	return ctx.Render(`server/file`, err)
}

func Editable(fileName string) (string, bool) {
	return config.FromFile().Sys.Editable(fileName)
}

func Playable(fileName string) (string, bool) {
	return config.FromFile().Sys.Playable(fileName)
}
