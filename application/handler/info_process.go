package handler

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/admpub/log"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/param"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"

	"github.com/nging-plugins/servermanager/application/library/system"
)

func processInfo(ctx context.Context, pid int32) (echo.H, error) {
	procs, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, err
	}
	cpuPercent, _ := procs.CPUPercentWithContext(ctx)
	memPercent, _ := procs.MemoryPercentWithContext(ctx)
	name, _ := procs.NameWithContext(ctx)
	cmdLine, _ := procs.CmdlineWithContext(ctx)
	exe, _ := procs.ExeWithContext(ctx)
	createTime, _ := procs.CreateTimeWithContext(ctx)
	row := echo.H{
		"name":           name,
		"cmd_line":       cmdLine,
		"exe":            exe,
		"created":        "",
		"cpu_percent":    cpuPercent,
		"memory_percent": memPercent,
	}
	if createTime > 0 {
		row["created"] = com.DateFormat(`Y-m-d H:i:s`, createTime/1000)
	}
	return row, nil
}

func ProcessInfo(ctx echo.Context) error {
	pid := ctx.Paramx(`pid`).Int32()
	row, err := processInfo(ctx, pid)
	data := ctx.Data()
	if err != nil {
		data.SetError(err)
	} else {
		data.SetData(row)
	}
	return ctx.JSON(data)
}

func ProcessKill(ctx echo.Context) error {
	pid := ctx.Paramx(`pid`).Int()
	err := com.CloseProcessFromPid(pid)
	data := ctx.Data()
	if err != nil {
		data.SetError(err)
	} else {
		data.SetData(nil)
	}
	return ctx.JSON(data)
}

type cacheProcess struct {
	processList          []*system.Process
	processLastQueryTime time.Time
	processQuering       bool
}

var cachedProcess cacheProcess
var processLock sync.RWMutex

func getCachedProc() cacheProcess {
	processLock.RLock()
	c := cachedProcess
	processLock.RUnlock()
	return c
}

func setCachedProc(c cacheProcess) {
	processLock.Lock()
	cachedProcess = c
	processLock.Unlock()
}

func ProcessList(ctx echo.Context) error {
	if ctx.Formx(`status`).Bool() {
		cp := getCachedProc()
		data := echo.H{`quering`: cp.processQuering}
		if !cp.processQuering {
			data.Set(`queryTime`, cp.processLastQueryTime.Format(param.DateTimeNormal))
		}
		return ctx.JSON(ctx.Data().SetData(data))
	}
	force := ctx.Formx(`force`).Bool()
	cp := getCachedProc()
	var err error
	var list []*system.Process
	var isCached bool
	if !cp.processQuering {
		if force || cp.processLastQueryTime.Before(time.Now().Add(-30*time.Minute)) {
			cp.processQuering = true
			setCachedProc(cp)
			go func() {
				stdCtx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
				defer cancel()
				list, err = system.ProcessList(stdCtx)
				if err != nil {
					log.Warn(err)
				}
				cp := getCachedProc()
				cp.processList = list
				cp.processQuering = false
				cp.processLastQueryTime = time.Now()
				setCachedProc(cp)
			}()
		} else {
			list = cp.processList
			isCached = true
		}
	}
	if force {
		return ctx.Redirect(backend.URLFor(`/server/processes`))
	}
	switch ctx.Form(`sort`) {
	case `cpu`:
		sortedList := system.ProcessListSortByCPUPercent(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `-cpu`:
		sortedList := system.ProcessListSortByCPUPercentReverse(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `mem`:
		sortedList := system.ProcessListSortByMemPercent(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `-mem`:
		sortedList := system.ProcessListSortByMemPercentReverse(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `thread`:
		sortedList := system.ProcessListSortByNumThreads(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `-thread`:
		sortedList := system.ProcessListSortByNumThreadsReverse(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `fd`:
		sortedList := system.ProcessListSortByNumFDs(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `-fd`:
		sortedList := system.ProcessListSortByNumFDsReverse(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `created`:
		sortedList := system.ProcessListSortByCreated(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `-created`:
		sortedList := system.ProcessListSortByCreatedReverse(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	case `-id`:
		sortedList := system.ProcessListSortByPidReverse(list)
		sort.Sort(sortedList)
		ctx.Set(`listData`, sortedList)
	default:
		ctx.Set(`listData`, list)
	}
	ctx.Set(`lastQueryTime`, cp.processLastQueryTime)
	ctx.Set(`isCached`, isCached)
	ctx.Set(`quering`, cp.processQuering)
	ctx.Set(`activeURL`, `/server/sysinfo`)
	if ctx.Formx(`partial`).Bool() {
		data := ctx.Data()
		if err != nil {
			return ctx.JSON(data.SetInfo(err.Error(), 0))
		}
		var partialBytes []byte
		partialBytes, err = ctx.Fetch(`server/processes_list_partial`, nil)
		data.SetData(echo.H{
			`html`: string(partialBytes),
		})
		return ctx.JSON(data)
	}
	return ctx.Render(`server/processes`, common.Err(ctx, err))
}
