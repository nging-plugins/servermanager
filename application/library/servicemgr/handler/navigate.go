package handler

import (
	"github.com/coscms/webcore/library/navigate"
	"github.com/webx-top/echo"
)

var leftNavigate = navigate.List{
	{
		Name:   echo.T(`重载系统服务后台`),
		Action: `system_service/daemon_reload`,
	},
	{
		Name:   echo.T(`重载系统服务`),
		Action: `system_service/reload`,
	},
	{
		Name:   echo.T(`重启系统服务`),
		Action: `system_service/restart`,
	},
	{
		Name:   echo.T(`停止系统服务`),
		Action: `system_service/stop`,
	},
	{
		Name:   echo.T(`启动系统服务`),
		Action: `system_service/start`,
	},
	{
		Name:   echo.T(`设置系统服务开机启动`),
		Action: `system_service/enable`,
	},
	// {
	// 	Name:   echo.T(`禁用系统服务`),
	// 	Action: `system_service/disable`,
	// },
	{
		Name:   echo.T(`查看系统服务配置文件`),
		Action: `system_service/list_files`,
	},
	{
		Name:   echo.T(`查看系统服务日志`),
		Action: `system_service/log`,
	},
	{
		Name:   echo.T(`清理系统服务日志`),
		Action: `system_service/log_clear`,
	},
}
