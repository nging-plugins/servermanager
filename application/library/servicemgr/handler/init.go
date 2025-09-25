package handler

import (
	routeRegistry "github.com/coscms/webcore/registry/route"
	"github.com/nging-plugins/servermanager/application/handler"
	"github.com/webx-top/echo"
)

func init() {
	handler.AddRouteRegister(registerRouteSystemService)
	handler.SetSystemServiceListQuerier(systemServiceList)
	handler.LeftNavigate.Children.Add(-1, leftNavigate...)
}

func registerRouteSystemService(r echo.RouteRegister) {
	metaHandler := routeRegistry.IRegister().MetaHandler
	g := r.Group(`/system_service`)
	g.Route(`GET,POST`, `/daemon_reload`, metaHandler(echo.H{`name`: `重载系统服务后台`}, systemServiceDaemonReload))
	g.Route(`GET,POST`, `/reload`, metaHandler(echo.H{`name`: `重载系统服务`}, systemServiceReload))
	g.Route(`GET,POST`, `/restart`, metaHandler(echo.H{`name`: `重启系统服务`}, systemServiceRestart))
	g.Route(`GET,POST`, `/stop`, metaHandler(echo.H{`name`: `停止系统服务`}, systemServiceStop))
	g.Route(`GET,POST`, `/start`, metaHandler(echo.H{`name`: `启动系统服务`}, systemServiceStart))
	g.Route(`GET,POST`, `/enable`, metaHandler(echo.H{`name`: `启用系统服务`}, systemServiceEnable))
	g.Route(`GET,POST`, `/disable`, metaHandler(echo.H{`name`: `禁用系统服务`}, systemServiceDisable))
	g.Route(`GET`, `/list_files`, metaHandler(echo.H{`name`: `查看系统服务配置文件`}, systemServiceListFiles))
	g.Route(`GET`, `/log`, metaHandler(echo.H{`name`: `查看系统服务日志`}, systemServiceLog))
	g.Route(`GET`, `/log_clear`, metaHandler(echo.H{`name`: `清理系统服务日志`}, systemServiceLogClear))
}
