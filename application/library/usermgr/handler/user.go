//go:build linux

package handler

import (
	"strings"

	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"

	usermgr "github.com/nging-plugins/servermanager/application/library/usermgr"
)

// SystemUserList shows the system user list page.
func SystemUserList(ctx echo.Context) error {
	client, err := usermgr.NewClient()
	if err != nil {
		return err
	}
	users, err := client.List(ctx)
	if err != nil {
		return err
	}
	ctx.Set(`listData`, users)
	return ctx.Render(`server/system_user`, common.Err(ctx, err))
}

// SystemUserAdd handles adding a new system user.
func SystemUserAdd(ctx echo.Context) error {
	var err error
	client, err := usermgr.NewClient()
	if err != nil {
		return err
	}
	if ctx.IsPost() {
		shell := resolveShell(ctx)
		u := &usermgr.User{
			Username: ctx.Form(`username`),
			Comment:  ctx.Form(`comment`),
			Shell:    shell,
			HomeDir:  ctx.Form(`homeDir`),
		}
		groupsStr := ctx.Form(`groups`)
		if groupsStr != `` {
			u.Groups = strings.Split(groupsStr, `,`)
		}
		password := ctx.Form(`password`)
		confirmPwd := ctx.Form(`confirmPassword`)
		if err = validateUserAddForm(ctx, u, password, confirmPwd); err != nil {
			goto END
		}
		err = client.Add(ctx, u, password)
		if err == nil {
			common.SendOk(ctx, ctx.T(`添加成功`))
			return ctx.Redirect(backend.URLFor(`/server/system_user`))
		}
	}

END:
	// Load shells for dropdown
	shells, shellErr := client.AvailableShells(ctx)
	if shellErr == nil {
		ctx.Set(`shells`, shells)
	}
	ctx.Set(`activeURL`, `/server/system_user`)
	return ctx.Render(`server/system_user_edit`, common.Err(ctx, err))
}

// SystemUserEdit handles editing a system user.
func SystemUserEdit(ctx echo.Context) error {
	username := ctx.Form(`username`)
	var err error
	client, err := usermgr.NewClient()
	if err != nil {
		return err
	}
	user, err := client.Get(ctx, username)
	if err != nil {
		return err
	}

	if ctx.IsPost() {
		shell := resolveShell(ctx)
		u := &usermgr.User{
			Comment: ctx.Form(`comment`),
			Shell:   shell,
			HomeDir: ctx.Form(`homeDir`),
		}
		groupsStr := ctx.Form(`groups`)
		if groupsStr != `` {
			u.Groups = strings.Split(groupsStr, `,`)
		}
		password := ctx.Form(`password`)
		confirmPwd := ctx.Form(`confirmPassword`)
		if password != `` && password != confirmPwd {
			err = ctx.NewError(code.InvalidParameter, ctx.T(`两次输入的密码不一致`)).SetZone(`confirmPassword`)
			goto END
		}
		err = client.Edit(ctx, username, u, password)
		if err == nil {
			common.SendOk(ctx, ctx.T(`修改成功`))
			return ctx.Redirect(backend.URLFor(`/server/system_user`))
		}
	} else {
		echo.StructToForm(ctx, user, ``, echo.LowerCaseFirstLetter)
		ctx.Request().Form().Set(`username`, username)
		if len(user.Groups) > 0 {
			ctx.Request().Form().Set(`groups`, strings.Join(user.Groups, `,`))
		}
		// Pass current shell value for template to detect "other" match
		ctx.Request().Form().Set(`_shell`, user.Shell)
	}

END:
	// Load shells for dropdown
	shells, shellErr := client.AvailableShells(ctx)
	if shellErr == nil {
		ctx.Set(`shells`, shells)
	}
	ctx.Set(`activeURL`, `/server/system_user`)
	return ctx.Render(`server/system_user_edit`, common.Err(ctx, err))
}

// resolveShell returns the actual shell path, handling the "other" option.
func resolveShell(ctx echo.Context) string {
	shell := ctx.Form(`shell`)
	if shell == `other` {
		if c := ctx.Form(`shell_custom`); c != `` {
			return c
		}
	}
	return shell
}

// SystemUserDelete handles deleting a system user.
func SystemUserDelete(ctx echo.Context) error {
	username := ctx.Form(`username`)
	removeHome := ctx.Form(`removeHome`) == `1`
	client, err := usermgr.NewClient()
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	err = client.Delete(ctx, username, removeHome)
	data := ctx.Data()
	if err != nil {
		data.SetError(err)
	} else {
		data.SetInfo(ctx.T(`删除成功`))
	}
	return ctx.JSON(data)
}

// SystemUserLock handles locking a system user account.
func SystemUserLock(ctx echo.Context) error {
	username := ctx.Form(`username`)
	client, err := usermgr.NewClient()
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	err = client.Lock(ctx, username)
	data := ctx.Data()
	if err != nil {
		data.SetError(err)
	} else {
		data.SetInfo(ctx.T(`已锁定`))
	}
	return ctx.JSON(data)
}

// SystemUserUnlock handles unlocking a system user account.
func SystemUserUnlock(ctx echo.Context) error {
	username := ctx.Form(`username`)
	client, err := usermgr.NewClient()
	if err != nil {
		return ctx.JSON(ctx.Data().SetError(err))
	}
	err = client.Unlock(ctx, username)
	data := ctx.Data()
	if err != nil {
		data.SetError(err)
	} else {
		data.SetInfo(ctx.T(`已解锁`))
	}
	return ctx.JSON(data)
}

func validateUserAddForm(ctx echo.Context, u *usermgr.User, password, confirmPwd string) error {
	if len(u.Username) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`用户名不能为空`)).SetZone(`username`)
	}
	if len(password) == 0 {
		return ctx.NewError(code.InvalidParameter, ctx.T(`密码不能为空`)).SetZone(`password`)
	}
	if password != confirmPwd {
		return ctx.NewError(code.InvalidParameter, ctx.T(`两次输入的密码不一致`)).SetZone(`confirmPassword`)
	}
	return nil
}
