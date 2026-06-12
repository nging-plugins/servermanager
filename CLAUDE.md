# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A plugin module for the [Nging](https://github.com/admpub/nging) webmaster toolbox, registered as module ID `server`. Adds server management capabilities to the Nging backend (based on `github.com/coscms/webcore`).

## API Conventions

### JSON Response Codes
- `Code == 1` — 成功
- `Code < 1` — 失败

## Build Commands

```bash
# Build
go build ./...

# Test a single package
go test ./application/library/servicemgr/

# Test with verbose output
go test -v ./application/library/config/

# Run all tests
go test ./...

# Generate DB schema code (from MySQL database)
# Requires: go install github.com/webx-top/db/cmd/dbgenerator@latest
bash gen_dbschema.sh
```

## Architecture

### Module Registration
- `module.go` — Plugin entry point. Registers routes, navigation, template paths, cron jobs, and install SQL via `module.Module{}` struct.
- `dashboard.go` — Registers dashboard blocks (CPU chart, command list) via init().
- `navigate.go` — Exports `LeftNavigate` tree, registered in `RegisterNavigate()`.
- `init_template.go` — Embeds template directory via `//go:embed` (build tag: `embedNgingPluginTemplate`).

### Route Structure
All routes under `/server` prefix, registered in `handler.RegisterRoute()`:
- `GET /sysinfo` — System information page
- `GET /netstat` — Network connections
- `GET,POST /processes` / `/process/:pid` / `/procskill/:pid` — Process management
- `GET,POST /service` — Systemd service management
- `GET,POST /hosts` — /etc/hosts file editor
- `GET,POST /command*` — Saved command CRUD
- `GET,POST /daemon_*` — Process daemon management
- `GET /cmd` — Command execution page
- WebSocket: `/cmdSendWS`, `/ptyWS`, `/dynamic`

### Key Libraries

| Package | Purpose |
|---|---|
| `application/handler/` | HTTP handlers and WebSocket endpoints |
| `application/model/` | DB models: Command, ForeverProcess |
| `application/dbschema/` | Auto-generated DB schema (nging_command, nging_forever_process) |
| `application/library/system/` | System monitoring via gopsutil/v4, time-series data, alerting |
| `application/library/config/` | Process daemon management (goforever) |
| `application/library/servicemgr/` | systemd D-Bus client (start/stop/restart/enable/disable) |
| `application/library/hosts/` | Hosts file reading/writing |
| `application/library/setup/` | Install SQL and role permission setup |
| `application/registry/` | Service control button registry |

### Data Flow
- **System monitoring**: gopsutil/v4 collects CPU/memory/disk/network/host data → `SystemInformation` struct → rendered in templates or streamed via WebSocket (`/dynamic`) as time-series data.
- **Command execution**: WebSocket/SockJS receives commands → `exec.Command` with timeout + output streaming → supports both local and SSH remote execution (via sshmanager plugin).
- **Process daemon**: goforever manages child processes → status persisted to `nging_forever_process` table → email alerts on unexpected exit.
- **Service management**: systemd D-Bus API via `github.com/coreos/go-systemd/v22` — list, start, stop, restart, enable, disable systemd units. Journalctl for logs.

### Cron Jobs
- `command` cron job type: executes saved commands on schedule. Runner in `handler.CommandJob()`.

### DB Schema Generation
Tables: `nging_command` (saved commands), `nging_forever_process` (daemon configs). Schema code auto-generated from MySQL via `dbgenerator`. Install SQL embedded in `application/library/setup/install.sql`.

## Template & Frontend Conventions

### Template Engine
Uses `github.com/webx-top/echo/middleware/render/standard` (standard Go `html/template`).

### Backend (Admin)
- Base template: `github.com/admpub/nging/template/backend` (Bootstrap 3)
- Assets: `github.com/admpub/nging/public/assets/backend`
- Plugin backend templates: `template/backend/` in this repo
- Plugin backend assets: `public/assets/backend/` in this repo (if any)

### Frontend (User-facing)
- Supports multi-theme. Each theme in a subfolder under `template/frontend/`.
- Default theme: `template/frontend/default` inherits from `github.com/admpub/webx/template/frontend/default` (Bootstrap 4)
- Frontend assets: `public/assets/frontend/<theme_name>/` — mapped to `github.com/admpub/webx/public/assets/frontend`
- Plugin frontend templates: `template/frontend/` in this repo
- Plugin frontend assets: `public/assets/frontend/` in this repo

### i18n (Multilingual)
- In templates: use `$.T` function. Patterns: `{{"文本"|$.T}}`, `{{`文本`|$.T}}`, `{{$.T "文本"}}`. Content left of `|` becomes last arg of `$.T`.
- In Go handler code: `ctx.T("文本")`
- Outside handler (package-level): `echo.T("文本")`

### Pagination
Uses `github.com/webx-top/db/lib/factory/pagination`. Sort field validation via `pagination.Sorts`:

```go
sorts := pagination.Sorts(ctx, `id`, `created`, `name`)
// Only allows sorting by id/created/name, rejects others
```

### Common Backend Handler Pattern
```go
func ListXxx(ctx echo.Context) error {
    m := model.NewXxx(ctx)
    cond := db.NewCompounds()
    err := m.ListPage(cond, func(r db.Result) db.Result {
        return r.OrderBy(`-id`)
    })
    ctx.Set(`listData`, m.Objects())
    return ctx.Render(`server/xxx`, common.Err(ctx, err))
}
```
