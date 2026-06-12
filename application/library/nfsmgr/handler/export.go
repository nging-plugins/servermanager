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
	"github.com/webx-top/echo/code"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"

	nfsmgr "github.com/nging-plugins/servermanager/application/library/nfsmgr"
)

// NFSExportList shows the NFS exports list page.
func NFSExportList(ctx echo.Context) error {
	client, err := nfsmgr.NewClient()
	if err != nil {
		return ctx.Data().SetError(err).JSON()
	}
	entries, err := client.ListExports(ctx)
	if err != nil {
		return ctx.Data().SetError(err).JSON()
	}
	ctx.Set(`listData`, entries)
	return ctx.Render(`server/nfs_export`, common.Err(ctx, err))
}

// NFSExportAdd handles adding a new NFS export entry.
func NFSExportAdd(ctx echo.Context) error {
	var err error
	if ctx.IsPost() {
		entry := &nfsmgr.ExportEntry{
			Path: ctx.Form(`path`),
		}
		if err = validateExportForm(ctx, entry); err != nil {
			goto END
		}
		client, err := nfsmgr.NewClient()
		if err != nil {
			goto END
		}
		entries, listErr := client.ListExports(ctx)
		if listErr != nil {
			err = listErr
			goto END
		}
		entries = append(entries, entry)
		err = client.WriteExports(ctx, entries)
		if err == nil {
			client.ReloadExports(ctx)
			common.SendOk(ctx, ctx.T(`操作成功`))
			return ctx.Redirect(backend.URLFor(`/server/nfs_export`))
		}
	}

END:
	ctx.Set(`activeURL`, `/server/nfs_export`)
	return ctx.Render(`server/nfs_export_edit`, common.Err(ctx, err))
}

// NFSExportEdit handles editing an NFS export entry.
func NFSExportEdit(ctx echo.Context) error {
	idx := ctx.Formx(`index`).Int()
	var err error
	client, err := nfsmgr.NewClient()
	if err != nil {
		return ctx.Data().SetError(err).JSON()
	}
	entries, err := client.ListExports(ctx)
	if err != nil {
		return ctx.Data().SetError(err).JSON()
	}
	if idx < 0 || idx >= len(entries) {
		return ctx.JSON(ctx.Data().SetError(ctx.NewError(code.InvalidParameter, ctx.T(`无效的索引`))))
	}
	entry := entries[idx]

	if ctx.IsPost() {
		entry.Path = ctx.Form(`path`)
		if err = validateExportForm(ctx, entry); err != nil {
			goto END
		}
		entries[idx] = entry
		err = client.WriteExports(ctx, entries)
		if err == nil {
			client.ReloadExports(ctx)
			common.SendOk(ctx, ctx.T(`修改成功`))
			return ctx.Redirect(backend.URLFor(`/server/nfs_export`))
		}
	} else {
		echo.StructToForm(ctx, entry, ``, echo.LowerCaseFirstLetter)
		ctx.Set(`clientList`, entry.Clients)
		ctx.Request().Form().Set(`index`, echo.String(idx))
	}

END:
	ctx.Set(`activeURL`, `/server/nfs_export`)
	return ctx.Render(`server/nfs_export_edit`, common.Err(ctx, err))
}

// NFSExportDelete handles deleting an NFS export entry.
func NFSExportDelete(ctx echo.Context) error {
	idx := ctx.Formx(`index`).Int()
	client, err := nfsmgr.NewClient()
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	entries, err := client.ListExports(ctx)
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	if idx < 0 || idx >= len(entries) {
		return ctx.JSON(ctx.Data().SetError(ctx.NewError(code.InvalidParameter, ctx.T(`无效的索引`))))
	}
	entries = append(entries[:idx], entries[idx+1:]...)
	err = client.WriteExports(ctx, entries)
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	client.ReloadExports(ctx)
	common.SendOk(ctx, ctx.T(`删除成功`))
	return ctx.Redirect(backend.URLFor(`/server/nfs_export`))
}

// NFSExportReload reloads NFS exports via exportfs -r.
func NFSExportReload(ctx echo.Context) error {
	client, err := nfsmgr.NewClient()
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	err = client.ReloadExports(ctx)
	data := ctx.Data()
	if err != nil {
		data.SetError(err)
	} else {
		data.SetInfo(ctx.T(`导出已重新加载`))
	}
	return ctx.JSON(data)
}

func validateExportForm(ctx echo.Context, entry *nfsmgr.ExportEntry) error {
	if len(entry.Path) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`导出路径不能为空`)).SetZone(`path`)
	}

	// Parse clients from form
	clientHosts := ctx.FormValues(`clientHost`)
	clientOptions := ctx.FormValues(`clientOptions`)
	entry.Clients = nil
	for i, host := range clientHosts {
		if len(host) == 0 {
			continue
		}
		c := nfsmgr.ExportClient{Host: host}
		if i < len(clientOptions) && len(clientOptions[i]) > 0 {
			c.Options = splitOptions(clientOptions[i])
		}
		entry.Clients = append(entry.Clients, c)
	}
	if len(entry.Clients) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`至少需要一个客户端`)).SetZone(`clientHost`)
	}

	// Parse common options
	commonOpts := ctx.Form(`options`)
	if len(commonOpts) > 0 {
		// Common options are applied to all clients
		opts := splitOptions(commonOpts)
		for i := range entry.Clients {
			entry.Clients[i].Options = append(opts, entry.Clients[i].Options...)
		}
	}
	return nil
}

func splitOptions(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
