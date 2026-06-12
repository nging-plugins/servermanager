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
	"github.com/coscms/webcore/library/navigate"
	"github.com/webx-top/echo"
)

var nfsLeftNavigate = navigate.List{
	{
		Display: true,
		Name:    echo.T("NFS管理"),
		Action:  "nfs",
		Icon:    "hdd-o",
	},
	{
		Display: false,
		Name:    echo.T("导出管理"),
		Action:  "nfs/export",
	},
	{
		Display: false,
		Name:    echo.T("添加导出"),
		Action:  "nfs/export_add",
	},
	{
		Display: false,
		Name:    echo.T("编辑导出"),
		Action:  "nfs/export_edit",
	},
	{
		Display: false,
		Name:    echo.T("删除导出"),
		Action:  "nfs/export_delete",
	},
	{
		Display: false,
		Name:    echo.T("重新加载导出"),
		Action:  "nfs/export_reload",
	},
	{
		Display: false,
		Name:    echo.T("挂载管理"),
		Action:  "nfs/mount",
	},
	{
		Display: false,
		Name:    echo.T("挂载NFS"),
		Action:  "nfs/mount_add",
	},
	{
		Display: false,
		Name:    echo.T("卸载"),
		Action:  "nfs/mount_umount",
	},
	{
		Display: false,
		Name:    echo.T("磁盘配额"),
		Action:  "nfs/quota",
	},
	{
		Display: false,
		Name:    echo.T("设置限额"),
		Action:  "nfs/quota_set",
	},
	{
		Display: false,
		Name:    echo.T("清除限额"),
		Action:  "nfs/quota_delete",
	},
}
