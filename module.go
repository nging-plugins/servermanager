package servermanager

import (
	"github.com/admpub/nging/v4/application/library/config/cmder"
	"github.com/admpub/nging/v4/application/library/module"

	"github.com/nging-plugins/servermanager/pkg/handler"
	pluginCmder "github.com/nging-plugins/servermanager/pkg/library/cmder"
	"github.com/nging-plugins/servermanager/pkg/library/setup"
)

const ID = `server`

var Module = module.Module{
	Startup: `daemon`,
	Cmder: map[string]cmder.Cmder{
		`daemon`: pluginCmder.New(),
	},
	TemplatePath: map[string]string{
		ID: `servermanager/template/backend`,
	},
	AssetsPath:    []string{},
	SQLCollection: setup.RegisterSQL,
	Navigate:      RegisterNavigate,
	Route:         handler.RegisterRoute,
}
