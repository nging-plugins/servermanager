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
	"os/exec"

	"github.com/webx-top/echo"

	"github.com/nging-plugins/servermanager/application/handler"
	nfsmgr "github.com/nging-plugins/servermanager/application/library/nfsmgr"
)

var existsExportFS bool

func init() {
	handler.LeftNavigate.Children.Add(-1, nfsLeftNavigate)
	handler.AddRouteRegister(registerIndex)
	// Mount + Quota routes don't need exportfs
	handler.AddRouteRegister(registerMountQuotaRoute)
	// Export routes require exportfs
	if _, err := exec.LookPath(`exportfs`); err == nil {
		existsExportFS = true
		handler.AddRouteRegister(registerExportRoute)
	}
}

func registerIndex(g echo.RouteRegister) {
	// NFS overview
	g.Route(`GET`, `/nfs`, NFSIndex)
}

func registerExportRoute(g echo.RouteRegister) {
	// Export management
	g.Route(`GET`, `/nfs_export`, NFSExportList)
	g.Route(`GET,POST`, `/nfs_export_add`, NFSExportAdd)
	g.Route(`GET,POST`, `/nfs_export_edit`, NFSExportEdit)
	g.Route(`GET,POST`, `/nfs_export_delete`, NFSExportDelete)
	g.Route(`POST`, `/nfs_export_reload`, NFSExportReload)
}

func registerMountQuotaRoute(g echo.RouteRegister) {
	// Mount management
	g.Route(`GET`, `/nfs_mount`, NFSMountList)
	g.Route(`GET,POST`, `/nfs_mount_add`, NFSMountAdd)
	g.Route(`POST`, `/nfs_mount_umount`, NFSMountUmount)
	// Quota management
	g.Route(`GET`, `/nfs_quota`, NFSQuota)
	g.Route(`GET,POST`, `/nfs_quota_set`, NFSQuotaSet)
}

// NFSIndex shows the NFS management overview page.
func NFSIndex(ctx echo.Context) error {
	var err error
	var status *nfsmgr.NFSStatus
	var mounts []*nfsmgr.MountEntry
	if existsExportFS {
		var client nfsmgr.Client
		client, err = nfsmgr.NewClient()
		if err != nil {
			return err
		}
		status, err = client.ServerStatus(ctx)
		if err != nil {
			ctx.Logger().Errorf(`failed to query NFS status: %v`, err)
		}
		mounts, err = client.ListMounts(ctx)
	}
	ctx.Set(`existsExportFS`, existsExportFS)
	ctx.Set(`nfsStatus`, status)
	ctx.Set(`mountList`, mounts)
	return ctx.Render(`server/nfs`, err)
}
