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
	"strconv"

	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"

	nfsmgr "github.com/nging-plugins/servermanager/application/library/nfsmgr"
)

// NFSQuota shows disk quota information for NFS export filesystems.
func NFSQuota(ctx echo.Context) error {
	client, err := nfsmgr.NewClient()
	if err != nil {
		return err
	}
	reports, err := client.ListQuota(ctx)
	if err != nil {
		return ctx.Render(`server/nfs_quota`, common.Err(ctx, err))
	}
	ctx.Set(`listData`, reports)
	ctx.Set(`activeURL`, `/server/nfs`)
	return ctx.Render(`server/nfs_quota`, common.Err(ctx, nil))
}

// NFSQuotaSet handles setting disk quota limits for a user.
func NFSQuotaSet(ctx echo.Context) error {
	var err error
	if ctx.IsPost() {
		limit := &nfsmgr.QuotaLimit{
			User:       ctx.Form(`user`),
			MountPoint: ctx.Form(`mountPoint`),
		}
		if err = validateQuotaForm(ctx, limit); err != nil {
			goto END
		}
		client, clientErr := nfsmgr.NewClient()
		if clientErr != nil {
			err = clientErr
			goto END
		}
		err = client.SetQuota(ctx, limit)
		if err == nil {
			common.SendOk(ctx, ctx.T(`限额设置成功`))
			return ctx.Redirect(backend.URLFor(`/server/nfs/quota`))
		}
	}

END:
	ctx.Set(`activeURL`, `/server/nfs`)
	return ctx.Render(`server/nfs_quota_set`, common.Err(ctx, err))
}

// NFSQuotaDelete clears disk quota limits for a user.
func NFSQuotaDelete(ctx echo.Context) error {
	user := ctx.Form(`user`)
	mountPoint := ctx.Form(`mountPoint`)
	if len(user) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`用户名不能为空`))
	}
	if len(mountPoint) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`挂载点不能为空`))
	}
	limit := &nfsmgr.QuotaLimit{
		User:       user,
		MountPoint: mountPoint,
	}
	client, err := nfsmgr.NewClient()
	if err != nil {
		return err
	}
	err = client.SetQuota(ctx, limit)
	if err == nil {
		common.SendOk(ctx, ctx.T(`配额已清除`))
	}
	return ctx.Redirect(backend.URLFor(`/server/nfs/quota`))
}

func validateQuotaForm(ctx echo.Context, limit *nfsmgr.QuotaLimit) error {
	if len(limit.User) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`用户名不能为空`)).SetZone(`user`)
	}
	if len(limit.MountPoint) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`挂载点不能为空`)).SetZone(`mountPoint`)
	}
	blockSoft := ctx.Form(`blockSoft`)
	blockHard := ctx.Form(`blockHard`)
	inodeSoft := ctx.Form(`inodeSoft`)
	inodeHard := ctx.Form(`inodeHard`)
	if len(blockSoft) > 0 {
		v, e := strconv.ParseUint(blockSoft, 10, 64)
		if e != nil {
			return ctx.NewError(code.InvalidParameter, ctx.T(`块软限制格式无效`)).SetZone(`blockSoft`)
		}
		limit.BlockSoft = v
	}
	if len(blockHard) > 0 {
		v, e := strconv.ParseUint(blockHard, 10, 64)
		if e != nil {
			return ctx.NewError(code.InvalidParameter, ctx.T(`块硬限制格式无效`)).SetZone(`blockHard`)
		}
		limit.BlockHard = v
	}
	if len(inodeSoft) > 0 {
		v, e := strconv.ParseUint(inodeSoft, 10, 64)
		if e != nil {
			return ctx.NewError(code.InvalidParameter, ctx.T(`Inode软限制格式无效`)).SetZone(`inodeSoft`)
		}
		limit.InodeSoft = v
	}
	if len(inodeHard) > 0 {
		v, e := strconv.ParseUint(inodeHard, 10, 64)
		if e != nil {
			return ctx.NewError(code.InvalidParameter, ctx.T(`Inode硬限制格式无效`)).SetZone(`inodeHard`)
		}
		limit.InodeHard = v
	}
	return nil
}
