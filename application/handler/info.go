/*
   Nging is a toolbox for webmasters
   Copyright (C) 2018-present  Wenhui Shen <swh@admpub.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package handler

import (
	"github.com/admpub/log"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/webx-top/echo"

	"github.com/nging-plugins/servermanager/application/library/system"
)

func Info(ctx echo.Context) error {
	var err error
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Warn(err)
	}
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Warn(err)
	}
	/*
		ioCounter, err := disk.IOCounters()
		if err != nil {
			log.Warn(err)
		}
	*/
	hostInfo, err := host.Info()
	if err != nil {
		log.Warn(err)
	}
	/*
		avgLoad, err := load.Avg()
		if err != nil {
			log.Warn(err)
		}
	*/
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		log.Warn(err)
	}
	swapMem, err := mem.SwapMemory()
	if err != nil {
		log.Warn(err)
	}
	if swapMem.UsedPercent == 0 {
		swapMem.UsedPercent = (float64(swapMem.Used) / float64(swapMem.Total)) * 100
	}
	netIOCounter, err := net.IOCounters(false)
	if err != nil {
		log.Warn(err)
	}
	cpuPercent, err := cpu.Percent(0, true)
	if err != nil {
		log.Warn(err)
	}
	info := &system.SystemInformation{
		CPU:        &system.CPUInformation{Percent: cpuPercent},
		Partitions: partitions,
		//DiskIO:         ioCounter,
		Host: hostInfo,
		//Load:       avgLoad,
		Memory: &system.MemoryInformation{Virtual: virtualMem, Swap: swapMem},
		NetIO:  netIOCounter,
		Go:     system.Status(),
	}
	if len(cpuInfo) > 0 {
		info.CPU.ModelName = cpuInfo[0].ModelName
	}
	info.CPU.Cores, _ = cpu.Counts(false)
	info.CPU.LogicalCores, _ = cpu.Counts(true)
	info.DiskUsages = make([]*disk.UsageStat, len(info.Partitions))
	for k, v := range info.Partitions {
		usageStat, err := disk.Usage(v.Mountpoint)
		if err != nil {
			log.Warn(err)
			continue
		}
		info.DiskUsages[k] = usageStat
	}

	info.Temp, _ = system.SensorsTemperatures()
	ctx.Data().SetData(info, 1)
	return ctx.Render(`server/sysinfo`, nil)
}
