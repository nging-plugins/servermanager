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
	"slices"
	"strings"

	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"

	nfsmgr "github.com/nging-plugins/servermanager/application/library/nfsmgr"
)

// Known checkbox option values for export options
var knownExportOpts = map[string]bool{
	`rw`:               true,
	`sync`:             true,
	`no_subtree_check`: true,
	`no_root_squash`:   true,
	`no_all_squash`:    true,
	`insecure`:         true,
	`crossmnt`:         true,
	`no_wdelay`:        true,
}

// NFSExportList shows the NFS exports list page.
func NFSExportList(ctx echo.Context) error {
	client, err := nfsmgr.NewClient()
	if err != nil {
		return err
	}
	entries, err := client.ListExports(ctx)
	if err != nil {
		return err
	}
	ctx.Set(`listData`, entries)
	ctx.Set(`activeURL`, `/server/nfs`)
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
		var client nfsmgr.Client
		client, err = nfsmgr.NewClient()
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
			return ctx.Redirect(backend.URLFor(`/server/nfs/export`))
		}
	}

END:
	if ctx.IsPost() {
		ctx.Request().Form().Set(`_exportOpts`, strings.Join(ctx.FormValues(`exportOpts`), `,`))
	}
	ctx.Set(`activeURL`, `/server/nfs`)
	return ctx.Render(`server/nfs_export_edit`, common.Err(ctx, err))
}

// NFSExportEdit handles editing an NFS export entry.
func NFSExportEdit(ctx echo.Context) error {
	ident := ctx.FormAny(`ident`, `path`)
	client, err := nfsmgr.NewClient()
	if err != nil {
		return err
	}
	entries, err := client.ListExports(ctx)
	if err != nil {
		return err
	}
	var foundIdx int = -1
	for i, e := range entries {
		if e.Path == ident {
			foundIdx = i
			break
		}
	}
	if foundIdx < 0 {
		return ctx.NewError(code.InvalidParameter, `导出配置不存在: %s`, ident)
	}
	entry := entries[foundIdx]

	if ctx.IsPost() {
		entry.Path = ctx.Form(`path`)
		if err = validateExportForm(ctx, entry); err != nil {
			goto END
		}
		entries[foundIdx] = entry
		err = client.WriteExports(ctx, entries)
		if err == nil {
			client.ReloadExports(ctx)
			common.SendOk(ctx, ctx.T(`修改成功`))
			return ctx.Redirect(backend.URLFor(`/server/nfs/export`))
		}
	} else {
		echo.StructToForm(ctx, entry, ``, echo.LowerCaseFirstLetter)
		ctx.Request().Form().Set(`ident`, ident)
		// Reconstruct form fields from options common to all clients
		if len(entry.Clients) > 0 {
			optCount := map[string]int{}
			cliIndex := map[string]map[int][]int{}
			for ci, c := range entry.Clients {
				seen := map[string]bool{}
				for oi, o := range c.Options {
					o = strings.TrimSpace(o)
					if len(o) == 0 {
						continue
					}
					if !seen[o] {
						seen[o] = true
						optCount[o]++
					}
					if _, ok := cliIndex[o]; !ok {
						cliIndex[o] = map[int][]int{
							ci: []int{oi},
						}
					} else {
						cliIndex[o][ci] = append(cliIndex[o][ci], oi)
					}
				}
			}
			var checkboxOpts, textOpts []string
			for o, n := range optCount {
				if n == len(entry.Clients) {
					if knownExportOpts[o] {
						checkboxOpts = append(checkboxOpts, o)
					} else {
						textOpts = append(textOpts, o)
					}
					for ci, indexes := range cliIndex[o] {
						for oi := range indexes {
							entry.Clients[ci].Options = slices.Delete(entry.Clients[ci].Options, oi, oi+1)
						}
					}
				}
			}
			if len(checkboxOpts) > 0 {
				ctx.Request().Form().Set(`_exportOpts`, strings.Join(checkboxOpts, `,`))
			}
			if len(textOpts) > 0 {
				ctx.Request().Form().Set(`options`, strings.Join(textOpts, `,`))
			}
		}
		ctx.Set(`clientList`, entry.Clients)
	}

END:
	if ctx.IsPost() {
		ctx.Request().Form().Set(`_exportOpts`, strings.Join(ctx.FormValues(`exportOpts`), `,`))
	}
	ctx.Set(`activeURL`, `/server/nfs`)
	return ctx.Render(`server/nfs_export_edit`, common.Err(ctx, err))
}

// NFSExportDelete handles deleting an NFS export entry.
func NFSExportDelete(ctx echo.Context) error {
	path := ctx.Form(`path`)
	client, err := nfsmgr.NewClient()
	if err != nil {
		return err
	}
	entries, err := client.ListExports(ctx)
	if err != nil {
		return err
	}
	var foundIdx int = -1
	for i, e := range entries {
		if e.Path == path {
			foundIdx = i
			break
		}
	}
	if foundIdx < 0 {
		return ctx.NewError(code.InvalidParameter, `导出配置不存在: %s`, path)
	}
	entries = append(entries[:foundIdx], entries[foundIdx+1:]...)
	err = client.WriteExports(ctx, entries)
	if err != nil {
		common.SendErr(ctx, err)
	} else {
		client.ReloadExports(ctx)
		common.SendOk(ctx, ctx.T(`删除成功`))
	}
	return ctx.Redirect(backend.URLFor(`/server/nfs/export`))
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

	// Collect common options from checkboxes + text field
	seen := map[string]bool{}
	var commonOpts []string
	for _, o := range ctx.FormValues(`exportOpts`) {
		o = strings.TrimSpace(o)
		if o != "" && !seen[o] {
			seen[o] = true
			commonOpts = append(commonOpts, o)
		}
	}
	for _, o := range splitOptions(ctx.Form(`options`)) {
		if !seen[o] {
			seen[o] = true
			commonOpts = append(commonOpts, o)
		}
	}
	if len(commonOpts) > 0 {
		// Common options are applied to all clients
		for i := range entry.Clients {
			entry.Clients[i].Options = append(commonOpts, entry.Clients[i].Options...)
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
