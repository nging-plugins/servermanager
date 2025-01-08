package handler

import (
	"github.com/coscms/webcore/library/navigate"
	"github.com/webx-top/echo"
)

var LeftNavigate = &navigate.Item{
	Display: true,
	Name:    echo.T(`服务器`),
	Action:  `server`,
	Icon:    `desktop`,
	Children: &navigate.List{
		{
			Display: true,
			Name:    echo.T(`服务器信息`),
			Action:  `sysinfo`,
		},
		{
			Display: false,
			Name:    echo.T(`服务器进程`),
			Action:  `processes`,
		},
		{
			Display: true,
			Name:    echo.T(`网络端口`),
			Action:  `netstat`,
		},
		{
			Display: true,
			Name:    echo.T(`执行命令`),
			Action:  `cmd`,
		},
		{
			Display: false,
			Name:    echo.T(`打开控制台`),
			Action:  `ptyWS`,
		},
		//快捷命令
		{
			Display: true,
			Name:    echo.T(`快捷命令`),
			Action:  `command`,
		},
		{
			Display: false,
			Name:    echo.T(`添加快捷命令`),
			Action:  `command_add`,
		},
		{
			Display: false,
			Name:    echo.T(`修改快捷命令`),
			Action:  `command_edit`,
		},
		{
			Display: false,
			Name:    echo.T(`删除快捷命令`),
			Action:  `command_delete`,
		},
		{
			Display: true,
			Name:    echo.T(`服务管理`),
			Action:  `service`,
		},
		{
			Display: false,
			Name:    echo.T(`查看服务日志`),
			Action:  `log/:category`,
		},
		{
			Display: true,
			Name:    echo.T(`hosts文件`),
			Action:  `hosts`,
		},
		{
			Display: false,
			Name:    echo.T(`查看进程详情`),
			Action:  `process/:pid`,
		},
		{
			Display: false,
			Name:    echo.T(`杀死进程`),
			Action:  `procskill/:pid`,
		},
		{
			Display: false,
			Name:    echo.T(`命令对话`),
			Action:  `cmdSend/*`,
		},
		{
			Display: false,
			Name:    echo.T(`发送命令`),
			Action:  `cmdSendWS`,
		},
		{
			Display: true,
			Name:    echo.T(`进程值守`),
			Action:  `daemon_index`,
		},
		{
			Display: false,
			Name:    echo.T(`进程值守日志`),
			Action:  `daemon_log`,
		},
		{
			Display: false,
			Name:    echo.T(`添加值守配置`),
			Action:  `daemon_add`,
		},
		{
			Display: false,
			Name:    echo.T(`修改值守配置`),
			Action:  `daemon_edit`,
		},
		{
			Display: false,
			Name:    echo.T(`删除值守配置`),
			Action:  `daemon_delete`,
		},
		{
			Display: false,
			Name:    echo.T(`重启值守`),
			Action:  `daemon_restart`,
		},
	},
}
