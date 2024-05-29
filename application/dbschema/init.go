// @generated Do not edit a file, which is automatically generated by the generator.

package dbschema

import (
	"github.com/webx-top/db/lib/factory"
)

var WithPrefix = func(tableName string) string {
	return "" + tableName
}

var DBI = factory.DefaultDBI

func init() {

	DBI.FieldsRegister(map[string]map[string]*factory.FieldInfo{"nging_command": {"command": {Name: "command", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "命令", GoType: "string", MyType: "", GoName: "Command"}, "created": {Name: "created", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "添加时间", GoType: "uint", MyType: "", GoName: "Created"}, "description": {Name: "description", DataType: "tinytext", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "说明", GoType: "string", MyType: "", GoName: "Description"}, "disabled": {Name: "disabled", DataType: "enum", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{"Y", "N"}, DefaultValue: "N", Comment: "是否禁用", GoType: "string", MyType: "", GoName: "Disabled"}, "env": {Name: "env", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "环境变量(一行一个，格式为：var1=val1)", GoType: "string", MyType: "", GoName: "Env"}, "id": {Name: "id", DataType: "int", Unsigned: true, PrimaryKey: true, AutoIncrement: true, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "ID", GoType: "uint", MyType: "", GoName: "Id"}, "name": {Name: "name", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 60, Options: []string{}, DefaultValue: "", Comment: "名称", GoType: "string", MyType: "", GoName: "Name"}, "remote": {Name: "remote", DataType: "enum", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{"Y", "N", "A"}, DefaultValue: "N", Comment: "是否(Y/N)执行远程SSH命令(A表示两者同时支持)", GoType: "string", MyType: "", GoName: "Remote"}, "ssh_account_id": {Name: "ssh_account_id", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "SSH账号ID(多个用逗号分隔)", GoType: "string", MyType: "", GoName: "SshAccountId"}, "updated": {Name: "updated", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "修改时间", GoType: "uint", MyType: "", GoName: "Updated"}, "work_directory": {Name: "work_directory", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "工作目录", GoType: "string", MyType: "", GoName: "WorkDirectory"}}, "nging_forever_process": {"args": {Name: "args", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "命令参数", GoType: "string", MyType: "", GoName: "Args"}, "command": {Name: "command", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 300, Options: []string{}, DefaultValue: "", Comment: "命令", GoType: "string", MyType: "", GoName: "Command"}, "created": {Name: "created", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "创建时间", GoType: "uint", MyType: "", GoName: "Created"}, "debug": {Name: "debug", DataType: "enum", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{"Y", "N"}, DefaultValue: "N", Comment: "DEBUG", GoType: "string", MyType: "", GoName: "Debug"}, "delay": {Name: "delay", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 30, Options: []string{}, DefaultValue: "", Comment: "延迟启动(例如1ms/1s/1m/1h)", GoType: "string", MyType: "", GoName: "Delay"}, "description": {Name: "description", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 500, Options: []string{}, DefaultValue: "", Comment: "说明", GoType: "string", MyType: "", GoName: "Description"}, "disabled": {Name: "disabled", DataType: "enum", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{"Y", "N"}, DefaultValue: "N", Comment: "是否禁用", GoType: "string", MyType: "", GoName: "Disabled"}, "enable_notify": {Name: "enable_notify", DataType: "tinyint", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "是否启用通知", GoType: "uint", MyType: "", GoName: "EnableNotify"}, "env": {Name: "env", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "环境变量", GoType: "string", MyType: "", GoName: "Env"}, "errfile": {Name: "errfile", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "错误记录文件", GoType: "string", MyType: "", GoName: "Errfile"}, "error": {Name: "error", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 300, Options: []string{}, DefaultValue: "", Comment: "错误信息", GoType: "string", MyType: "", GoName: "Error"}, "id": {Name: "id", DataType: "int", Unsigned: true, PrimaryKey: true, AutoIncrement: true, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "ID", GoType: "uint", MyType: "", GoName: "Id"}, "lastrun": {Name: "lastrun", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "上次运行时间", GoType: "uint", MyType: "", GoName: "Lastrun"}, "log_charset": {Name: "log_charset", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 10, Options: []string{}, DefaultValue: "", Comment: "日志字符集", GoType: "string", MyType: "", GoName: "LogCharset"}, "logfile": {Name: "logfile", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "日志记录文件", GoType: "string", MyType: "", GoName: "Logfile"}, "name": {Name: "name", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 60, Options: []string{}, DefaultValue: "", Comment: "名称", GoType: "string", MyType: "", GoName: "Name"}, "notify_email": {Name: "notify_email", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "通知人列表", GoType: "string", MyType: "", GoName: "NotifyEmail"}, "options": {Name: "options", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 1000, Options: []string{}, DefaultValue: "", Comment: "其它选项值(JSON)", GoType: "string", MyType: "", GoName: "Options"}, "pid": {Name: "pid", DataType: "int", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: -0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "PID", GoType: "int", MyType: "", GoName: "Pid"}, "pidfile": {Name: "pidfile", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "PID记录文件", GoType: "string", MyType: "", GoName: "Pidfile"}, "ping": {Name: "ping", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 30, Options: []string{}, DefaultValue: "", Comment: "心跳时间(例如1ms/1s/1m/1h)", GoType: "string", MyType: "", GoName: "Ping"}, "respawn": {Name: "respawn", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "1", Comment: "重试次数(进程被外部程序结束后自动启动)", GoType: "uint", MyType: "", GoName: "Respawn"}, "status": {Name: "status", DataType: "enum", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{"started", "running", "stopped", "restarted", "exited", "killed", "idle"}, DefaultValue: "idle", Comment: "进程运行状态", GoType: "string", MyType: "", GoName: "Status"}, "uid": {Name: "uid", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "添加人ID", GoType: "uint", MyType: "", GoName: "Uid"}, "updated": {Name: "updated", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "修改时间", GoType: "uint", MyType: "", GoName: "Updated"}, "user": {Name: "user", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 225, Options: []string{}, DefaultValue: "", Comment: "用户名", GoType: "string", MyType: "", GoName: "User"}, "workdir": {Name: "workdir", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "工作目录", GoType: "string", MyType: "", GoName: "Workdir"}}})

	DBI.ColumnsRegister(map[string][]string{"nging_command": {"id", "name", "description", "command", "work_directory", "env", "created", "updated", "disabled", "remote", "ssh_account_id"}, "nging_forever_process": {"id", "uid", "name", "command", "workdir", "env", "args", "pidfile", "logfile", "errfile", "log_charset", "respawn", "delay", "ping", "pid", "status", "debug", "disabled", "created", "updated", "error", "lastrun", "description", "user", "options", "enable_notify", "notify_email"}})

	DBI.ModelsRegister(factory.ModelInstancers{`NgingCommand`: factory.NewMI("nging_command", func(connID int) factory.Model { return &NgingCommand{base: *factory.NewBase(connID)} }, "快捷命令"), `NgingForeverProcess`: factory.NewMI("nging_forever_process", func(connID int) factory.Model { return &NgingForeverProcess{base: *factory.NewBase(connID)} }, "持久进程")})

}
