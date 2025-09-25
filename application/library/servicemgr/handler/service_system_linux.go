package handler

import (
	"bytes"
	"io"
	"strings"

	"github.com/webx-top/com"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"

	"github.com/nging-plugins/servermanager/application/library/servicemgr"
)

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
