package handler

import (
	"bytes"
	"io"
	"strings"

	"github.com/webx-top/com"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"

	"github.com/coscms/webcore/library/navigate"
	routeRegistry "github.com/coscms/webcore/registry/route"
	"github.com/nging-plugins/servermanager/application/handler"
	"github.com/nging-plugins/servermanager/application/library/servicemgr"
)

var LeftNavigate = navigate.List{
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
		Name:   echo.T(`启用系统服务`),
		Action: `system_service/enable`,
	},
	{
		Name:   echo.T(`禁用系统服务`),
		Action: `system_service/disable`,
	},
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

func init() {
	handler.AddRouteRegister(registerRouteSystemService)
	handler.SetSystemServiceListQuerier(systemServiceList)
	handler.LeftNavigate.Children.Add(-1, LeftNavigate...)
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

func systemServiceDaemonReload(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	err = client.Reload(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	return ctx.JSON(data.SetInfo(ctx.T("服务后台重载成功"), code.Success.Int()))
}

func systemServiceList(ctx echo.Context) error {
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return err
	}
	state := ctx.Formx("state").String()
	var states []string
	if state != "" && com.StrIsAlphaNumeric(state) {
		states = []string{state}
	}
	var patterns []string
	name := ctx.FormAnyx("name", "q").String()
	if name != "" {
		if err := validateServiceName(ctx, name); err != nil {
			return err
		}
		name = servicemgr.GetServiceName(name)
		patterns = append(patterns, name)
	}
	list, err := client.List(ctx, states, patterns)
	if err != nil {
		return err
	}
	ctx.Set(`systemServiceList`, list)
	return err
}

func systemServiceListFiles(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name = servicemgr.GetServiceName(name)
	list, err := client.ListFiles(ctx, []string{name})
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	data.SetData(echo.H{`list`: servicemgr.ServiceConfFileWithContent(list)})
	return ctx.JSON(data)
}

func validateServiceName(ctx echo.Context, name string) error {
	if strings.ContainsAny(name, "\n\t\r'\"") || com.IllegalFilePath(name) {
		return ctx.NewError(code.InvalidParameter, "Invalid service name").SetZone(`name`)
	}
	return nil
}

func getServiceName(ctx echo.Context) (string, error) {
	name := ctx.Formx("name").String()
	if name == "" {
		return "", ctx.NewError(code.InvalidParameter, "Missing service name").SetZone(`name`)
	}
	err := validateServiceName(ctx, name)
	return name, err
}

func systemServiceRestart(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	err = client.Restart(ctx, name)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	return ctx.JSON(data.SetInfo(ctx.T("服务重启成功"), code.Success.Int()))
}

func systemServiceReload(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	err = client.ReloadUnit(ctx, name)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	return ctx.JSON(data.SetInfo(ctx.T("服务重启成功"), code.Success.Int()))
}

func systemServiceStop(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	err = client.Stop(ctx, name)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	return ctx.JSON(data.SetInfo(ctx.T("服务停止成功"), code.Success.Int()))
}

func systemServiceStart(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	err = client.Start(ctx, name)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	return ctx.JSON(data.SetInfo(ctx.T("服务启动成功"), code.Success.Int()))
}

func systemServiceEnable(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	err = client.Enable(ctx, name)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	return ctx.JSON(data.SetInfo(ctx.T("服务启用成功"), code.Success.Int()))
}

func systemServiceDisable(ctx echo.Context) error {
	data := ctx.Data()
	client, err := servicemgr.NewClient(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	err = client.Disable(ctx, name)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	return ctx.JSON(data.SetInfo(ctx.T("服务禁用成功"), code.Success.Int()))
}

func systemServiceLog(ctx echo.Context) error {
	data := ctx.Data()
	name, err := getServiceName(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	buf := bytes.NewBuffer(nil)
	lines := ctx.Formx("lastLines").Uint()
	if lines == 0 {
		lines = 100
	} else if lines > 500 {
		lines = 500
	}
	err = servicemgr.ServiceLogWithRows(ctx, name, lines, func(rd io.Reader) error {
		_, err := io.Copy(buf, rd)
		return err
	}, false)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	data.SetData(echo.H{`content`: buf.String()})
	return ctx.JSON(data)
}

func systemServiceLogClear(ctx echo.Context) error {
	data := ctx.Data()
	err := servicemgr.ServiceLogClear(ctx)
	if err != nil {
		return ctx.JSON(data.SetError(err))
	}
	data.SetInfo(ctx.T("服务日志清理成功"), code.Success.Int())
	return ctx.JSON(data)
}
