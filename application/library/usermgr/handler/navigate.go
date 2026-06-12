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

var usermgrLeftNavigate = &navigate.Item{
	Display: true,
	Name:    echo.T("系统用户"),
	Action:  "system_user",
	Icon:    "users",
	Children: &navigate.List{
		{
			Display: true,
			Name:    echo.T("用户列表"),
			Action:  "system_user",
		},
		{
			Display: false,
			Name:    echo.T("添加用户"),
			Action:  "system_user_add",
		},
		{
			Display: false,
			Name:    echo.T("编辑用户"),
			Action:  "system_user_edit",
		},
		{
			Display: false,
			Name:    echo.T("删除用户"),
			Action:  "system_user_delete",
		},
		{
			Display: false,
			Name:    echo.T("锁定用户"),
			Action:  "system_user_lock",
		},
		{
			Display: false,
			Name:    echo.T("解锁用户"),
			Action:  "system_user_unlock",
		},
	},
}
