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

// NFSMountList shows the list of mounted NFS shares.
func NFSMountList(ctx echo.Context) error {
	client, err := nfsmgr.NewClient()
	if err != nil {
		return ctx.Data().SetError(err).JSON()
	}
	mounts, err := client.ListMounts(ctx)
	if err != nil {
		return ctx.Data().SetError(err).JSON()
	}
	ctx.Set(`listData`, mounts)
	return ctx.Render(`server/nfs_mount`, common.Err(ctx, err))
}

// NFSMountAdd handles mounting a new NFS share.
func NFSMountAdd(ctx echo.Context) error {
	var err error
	if ctx.IsPost() {
		entry := &nfsmgr.MountEntry{
			Server:     ctx.Form(`server`),
			Remote:     ctx.Form(`remote`),
			MountPoint: ctx.Form(`mountPoint`),
			Type:       ctx.Form(`type`),
		}
		if entry.Type == "" {
			entry.Type = "nfs4"
		}
		if err = validateMountForm(ctx, entry); err != nil {
			goto END
		}
		opts := ctx.Form(`options`)
		if len(opts) > 0 {
			entry.Options = strings.Split(opts, ",")
		}
		client, clientErr := nfsmgr.NewClient()
		if clientErr != nil {
			err = clientErr
			goto END
		}
		err = client.Mount(ctx, entry)
		if err == nil {
			common.SendOk(ctx, ctx.T(`挂载成功`))
			return ctx.Redirect(backend.URLFor(`/server/nfs_mount`))
		}
	}

END:
	ctx.Set(`activeURL`, `/server/nfs_mount`)
	return ctx.Render(`server/nfs_mount_add`, common.Err(ctx, err))
}

// NFSMountUmount unmounts an NFS share.
func NFSMountUmount(ctx echo.Context) error {
	mountPoint := ctx.Form(`mountPoint`)
	if len(mountPoint) == 0 {
		return ctx.JSON(ctx.Data().SetError(ctx.NewError(code.InvalidParameter, ctx.T(`挂载点不能为空`))))
	}
	client, err := nfsmgr.NewClient()
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	err = client.Unmount(ctx, mountPoint)
	data := ctx.Data()
	if err != nil {
		data.SetError(err)
	} else {
		data.SetInfo(ctx.T(`卸载成功`))
	}
	return ctx.JSON(data)
}

func validateMountForm(ctx echo.Context, entry *nfsmgr.MountEntry) error {
	if len(entry.Server) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`服务器地址不能为空`)).SetZone(`server`)
	}
	if len(entry.Remote) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`远程路径不能为空`)).SetZone(`remote`)
	}
	if len(entry.MountPoint) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`本地挂载点不能为空`)).SetZone(`mountPoint`)
	}
	return nil
}
