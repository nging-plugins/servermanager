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

package system

import (
	"context"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"github.com/webx-top/com"
	"github.com/webx-top/echo/defaults"
	"github.com/webx-top/echo/param"

	"github.com/admpub/log"
	"github.com/coscms/webcore/library/cron"
	"github.com/coscms/webcore/library/msgbox"
	"github.com/coscms/webcore/registry/alert"
)

var (
	mutext         sync.Mutex
	realTimeStatus *RealTimeStatus
	// CancelRealTimeStatusCollection 取消实时状态搜集
	CancelRealTimeStatusCollection = func() {}
)

func init() {
	alert.Topics.Add(`systemStatus`, `系统状态`)
}

var poolRealTimeStatusLite = sync.Pool{
	New: func() any {
		return &RealTimeStatusLite{pooled: true}
	},
}

// RealTimeStatusObject 实时状态
func RealTimeStatusObject(n ...int) *RealTimeStatusLite {
	r := poolRealTimeStatusLite.Get().(*RealTimeStatusLite)
	if !r.pooled {
		r.pooled = true
	}
	if len(n) == 0 || n[0] <= 0 {
		return r.CopyFrom(realTimeStatus)
	}
	max := n[0]
	r.CPU = realTimeStatus.CPU.GetTruncate(max)
	r.Mem = realTimeStatus.Mem.GetTruncate(max)
	r.Net = NetIOTimeSeries{
		BytesSent:   realTimeStatus.Net.BytesSent.GetTruncate(max),
		BytesRecv:   realTimeStatus.Net.BytesRecv.GetTruncate(max),
		PacketsSent: realTimeStatus.Net.PacketsSent.GetTruncate(max),
		PacketsRecv: realTimeStatus.Net.PacketsRecv.GetTruncate(max),
	}
	r.Temp = make(map[string]TimeSeries, len(realTimeStatus.Temp))
	for key, value := range realTimeStatus.Temp {
		r.Temp[key] = value.GetTruncate(max)
	}
	return r
}

// RealTimeStatusIsListening 是否正在监听实时状态
func RealTimeStatusIsListening() bool {
	return realTimeStatus != nil && realTimeStatus.status == `started`
}

// ListenRealTimeStatus 监听实时状态
func ListenRealTimeStatus(cfg *Settings) {
	mutext.Lock()
	defer mutext.Unlock()
	interval := time.Second * 2
	max := 80
	if RealTimeStatusIsListening() {
		CancelRealTimeStatusCollection()
		realTimeStatus.SetSettings(cfg, interval, max)
	} else {
		realTimeStatus = NewRealTimeStatus(cfg, interval, max)
	}

	msgbox.Info(`System Monitor`, `Starting collect server status`)

	ctx, cancel := context.WithCancel(context.Background())
	go realTimeStatus.Listen(ctx)
	CancelRealTimeStatusCollection = func() {
		if RealTimeStatusIsListening() {
			cancel()
		}
	}
}

// NewRealTimeStatus 创建实时状态数据结构
func NewRealTimeStatus(cfg *Settings, interval time.Duration, maxSize int) *RealTimeStatus {
	r := &RealTimeStatus{
		max:        maxSize,
		interval:   interval,
		CPU:        TimeSeries{},
		Mem:        TimeSeries{},
		Net:        NewNetIOTimeSeries(),
		Temp:       map[string]TimeSeries{},
		reportTime: map[string]time.Time{},
	}
	return r.SetSettings(cfg, interval, maxSize)
}

// NewNetIOTimeSeries 创建网络IO时序数据结构
func NewNetIOTimeSeries() NetIOTimeSeries {
	return NetIOTimeSeries{
		lastBytesSent:   LastTimeValue{},
		lastBytesRecv:   LastTimeValue{},
		lastPacketsSent: LastTimeValue{},
		lastPacketsRecv: LastTimeValue{},
		BytesSent:       TimeSeries{},
		BytesRecv:       TimeSeries{},
		PacketsSent:     TimeSeries{},
		PacketsRecv:     TimeSeries{},
	}
}

// LastTimeValue 上次时间的状态值
type LastTimeValue struct {
	Time  time.Time
	Value float64
}

// NetIOTimeSeries 网络IO时序数据结构
type NetIOTimeSeries struct {
	lastBytesSent   LastTimeValue
	lastBytesRecv   LastTimeValue
	lastPacketsSent LastTimeValue
	lastPacketsRecv LastTimeValue

	BytesSent   TimeSeries
	BytesRecv   TimeSeries
	PacketsSent TimeSeries
	PacketsRecv TimeSeries
}

type RealTimeStatusLite struct {
	CPU    TimeSeries
	Mem    TimeSeries
	Net    NetIOTimeSeries
	Temp   map[string]TimeSeries
	pooled bool
}

func (r *RealTimeStatusLite) Release() {
	if r.pooled {
		r.CPU = nil
		r.Mem = nil
		r.Net.BytesSent = nil
		r.Net.BytesRecv = nil
		r.Net.PacketsSent = nil
		r.Net.PacketsRecv = nil
		r.Temp = nil
		poolRealTimeStatusLite.Put(r)
	}
}

func (r *RealTimeStatusLite) CopyFrom(f *RealTimeStatus) *RealTimeStatusLite {
	r.CPU = f.CPU
	r.Mem = f.CPU
	r.Net = f.Net
	r.Temp = f.Temp
	return r
}

// RealTimeStatus 实时状态数据结构
type RealTimeStatus struct {
	max         int
	interval    time.Duration
	CPU         TimeSeries
	Mem         TimeSeries
	Net         NetIOTimeSeries
	Temp        map[string]TimeSeries
	settings    *Settings
	reportEmail []string
	reportTime  map[string]time.Time
	status      string
	lock        sync.RWMutex
}

// Listen 监听
func (r *RealTimeStatus) Listen(ctx context.Context) *RealTimeStatus {
	r.status = `started`
	info := &DynamicInformation{}
	t := time.NewTicker(r.interval)
	defer func() {
		if e := recover(); e != nil {
			log.Errorf(`RealTimeStatus: %v`, e)
		}
		t.Stop()
	}()
	for {
		select {
		case <-ctx.Done():
			msgbox.Warn(`System Monitor`, `Exit server real-time status collection`)
			r.status = `stoped`
			return r
		case <-t.C:
			info.NetMemoryCPU()
			if len(info.CPUPercent) > 0 {
				r.CPUAdd(info.CPUPercent[0])
			} else {
				r.CPUAdd(0)
			}
			r.MemAdd(info.Memory.Virtual.UsedPercent)
			if len(info.NetIO) > 0 {
				r.NetAdd(info.NetIO[0])
			}
			info.TemperatureStat()
			if len(info.Temp) > 0 {
				r.TempAdd(info.Temp)
			}
			//log.Info(`Collect server status`)
		}
	}
}

func checkAndSendAlarm(r *RealTimeStatus, value float64, typ string, subType ...string) {
	if r == nil || r.settings == nil {
		return
	}
	if !r.settings.AlarmOn {
		return
	}
	switch typ {
	case `CPU`:
		if r.settings.AlarmThreshold.CPU > 0 && r.settings.AlarmThreshold.CPU < value {
			r.sendAlarm(r.settings.AlarmThreshold.CPU, value, typ)
			return
		}
	case `Temp`:
		if r.settings.AlarmThreshold.Temp > 0 && r.settings.AlarmThreshold.Temp < value {
			r.sendAlarm(r.settings.AlarmThreshold.Temp, value, typ, subType...)
			return
		}
	case `Mem`:
		if r.settings.AlarmThreshold.Memory > 0 && r.settings.AlarmThreshold.Memory < value {
			r.sendAlarm(r.settings.AlarmThreshold.Memory, value, typ)
			return
		}
	}
}

type alarmContent struct {
	title    string
	hostname string
	typeName string
	statType string
	subType  string
	value    string
}

func (a *alarmContent) genEmailContent() string {
	var content string
	if a.statType == `Temp` {
		content = a.subType + a.typeName + `: ` + a.value + `摄氏度`
	} else {
		content = a.typeName + `使用率: ` + a.value + `%`
	}
	return content
}

func (a *alarmContent) genMarkdownContent() string {
	var content string
	if a.statType == `Temp` {
		content = `**` + a.subType + a.typeName + `**: ` + a.value + `摄氏度`
	} else {
		content = `**` + a.typeName + `使用率**: ` + a.value + `%`
	}
	return content
}

func (a *alarmContent) EmailContent(_ param.Store) []byte {
	return com.Str2bytes(`<h1>` + a.title + `</h1><p>主机名: ` + a.hostname + `<br />` + a.genEmailContent() + `<br />时间: ` + time.Now().Format(time.RFC3339) + `<br /></p>`)
}

func (a *alarmContent) MarkdownContent(_ param.Store) []byte {
	return com.Str2bytes(`### ` + a.title + "\n" + `**主机名**: ` + a.hostname + "\n" + a.genMarkdownContent() + "\n" + `**时间**: ` + time.Now().Format(time.RFC3339) + "\n")
}

func (r *RealTimeStatus) sendAlarm(alarmThreshold, value float64, typ string, subType ...string) *RealTimeStatus {
	now := time.Now()
	var (
		reportTime time.Time
		ok         bool
	)
	if r.reportTime != nil {
		reportTime, ok = r.reportTime[typ]
	}
	if ok && !reportTime.IsZero() && now.Sub(reportTime) < time.Minute*5 { // 连续5分钟达到阀值时发邮件告警
		return nil
	}
	if r.reportTime == nil {
		r.reportTime = map[string]time.Time{
			typ: now,
		}
	} else {
		r.reportTime[typ] = now
	}
	var typeName, title string
	hostname, _ := os.Hostname()
	switch typ {
	case `CPU`:
		typeName = `CPU`
		title = fmt.Sprintf(`【`+hostname+`】`+typeName+`使用率超出%v%%`, alarmThreshold)
	case `Temp`:
		if len(subType) < 1 {
			return nil
		}
		typeName = `温度`
		title = fmt.Sprintf(`【`+hostname+`】`+subType[0]+typeName+`超过%v摄氏度`, alarmThreshold)
	case `Mem`:
		typeName = `内存`
		title = fmt.Sprintf(`【`+hostname+`】`+typeName+`使用率超出%v%%`, alarmThreshold)
	default:
		return nil
	}
	ct := &alarmContent{
		title:    title,
		hostname: hostname,
		typeName: typeName,
		statType: typ,
		subType:  ``,
		value:    fmt.Sprint(value),
	}
	if len(subType) > 0 {
		ct.subType = subType[0]
	}
	alertData := &alert.AlertData{
		Title:   title,
		Content: ct,
		Data:    param.Store{},
	}
	ctx := defaults.NewMockContext()
	if err := alert.SendTopic(ctx, `systemStatus`, alertData); err != nil {
		log.Warn(`alert.SendTopic: `, err)
	}
	if len(r.reportEmail) == 0 {
		return r
	}
	content := ct.EmailContent(alertData.Data)
	var cc []string
	if len(r.reportEmail) > 1 {
		cc = r.reportEmail[1:]
	}
	err := cron.SendMail(r.reportEmail[0], `administrator`, title, content, cc...)
	if err != nil {
		log.Errorf(`failed to cron.SendMail: %v`, err)
	}
	return r
}

func (r *RealTimeStatus) SetSettings(c *Settings, interval time.Duration, max int) *RealTimeStatus {
	r.settings = c
	var reportEmail []string
	if c != nil {
		if len(c.ReportEmail) > 0 {
			for _, email := range strings.Split(c.ReportEmail, "\n") {
				email = strings.TrimSpace(email)
				if len(email) == 0 {
					continue
				}
				reportEmail = append(reportEmail, email)
			}
		}
	}
	r.reportEmail = reportEmail
	r.interval = interval
	r.max = max
	return r
}

func (r *RealTimeStatus) CPUAdd(y float64) *RealTimeStatus {
	if r.max <= 0 {
		return r
	}
	r.lock.Lock()
	checkAndSendAlarm(r, y, `CPU`)
	r.CPU.Truncate(r.max - 1)
	r.CPU.Add(NewXY(y))
	r.lock.Unlock()
	return r
}

func (r *RealTimeStatus) TempAdd(ts []sensors.TemperatureStat) *RealTimeStatus {
	if r.max <= 0 {
		return r
	}
	r.lock.Lock()
	if r.Temp == nil {
		r.Temp = map[string]TimeSeries{}
	}
	for _, temp := range ts {
		checkAndSendAlarm(r, temp.Temperature, `Temp`, temp.SensorKey)
		_temp, ok := r.Temp[temp.SensorKey]
		if !ok {
			r.Temp[temp.SensorKey] = []XY{NewXY(temp.Temperature)}
			continue
		}
		_temp.Truncate(r.max - 1)
		_temp.Add(NewXY(temp.Temperature))
		r.Temp[temp.SensorKey] = _temp
	}
	r.lock.Unlock()
	return r
}

func (r *RealTimeStatus) MemAdd(y float64) *RealTimeStatus {
	if r.max <= 0 {
		return r
	}
	r.lock.Lock()
	checkAndSendAlarm(r, y, `Mem`)
	r.Mem.Truncate(r.max - 1)
	r.Mem.Add(NewXY(y))
	r.lock.Unlock()
	return r
}

func (r *RealTimeStatus) NetAdd(stat net.IOCountersStat) *RealTimeStatus {
	if r.max <= 0 {
		return r
	}
	r.lock.Lock()
	now := time.Now()
	r.Net.BytesRecv.Truncate(r.max - 1)
	n := float64(stat.BytesRecv)
	var speed float64
	if !r.Net.lastBytesRecv.Time.IsZero() {
		speed = (n - r.Net.lastBytesRecv.Value) / now.Sub(r.Net.lastBytesRecv.Time).Seconds()
		speed = math.Ceil(speed)
	} else {
		speed = 0
	}
	r.Net.BytesRecv.Add(NewXY(speed))
	r.Net.lastBytesRecv.Time = now
	r.Net.lastBytesRecv.Value = n

	r.Net.BytesSent.Truncate(r.max - 1)
	n = float64(stat.BytesSent)
	if !r.Net.lastBytesSent.Time.IsZero() {
		speed = (n - r.Net.lastBytesSent.Value) / now.Sub(r.Net.lastBytesSent.Time).Seconds()
		speed = math.Ceil(speed)
	} else {
		speed = 0
	}
	r.Net.BytesSent.Add(NewXY(speed))
	r.Net.lastBytesSent.Time = now
	r.Net.lastBytesSent.Value = n

	r.Net.PacketsRecv.Truncate(r.max - 1)
	n = float64(stat.PacketsRecv)
	if !r.Net.lastPacketsRecv.Time.IsZero() {
		speed = (n - r.Net.lastPacketsRecv.Value) / now.Sub(r.Net.lastPacketsRecv.Time).Seconds()
		speed = math.Ceil(speed)
	} else {
		speed = 0
	}
	r.Net.PacketsRecv.Add(NewXY(speed))
	r.Net.lastPacketsRecv.Time = now
	r.Net.lastPacketsRecv.Value = n

	r.Net.PacketsSent.Truncate(r.max - 1)
	n = float64(stat.PacketsSent)
	if !r.Net.lastPacketsSent.Time.IsZero() {
		speed = (n - r.Net.lastPacketsSent.Value) / now.Sub(r.Net.lastPacketsSent.Time).Seconds()
		speed = math.Ceil(speed)
	} else {
		speed = 0
	}
	r.Net.PacketsSent.Add(NewXY(speed))
	r.Net.lastPacketsSent.Time = now
	r.Net.lastPacketsSent.Value = n
	r.lock.Unlock()
	return r
}

type (
	TimeSeries []XY
	XY         [2]interface{}
)

func NewXY(y float64) XY {
	x := time.Now().UnixNano() / 1e6 //毫秒
	return XY{x, y}
}

func (t *TimeSeries) Truncate(max int) {
	length := len(*t)
	if length > max {
		*t = slices.Clone((*t)[length-max:])
	}
}

func (t TimeSeries) GetTruncate(max int) TimeSeries {
	length := len(t)
	if length > max {
		return slices.Clone(t[length-max:])
	}
	return t
}

func (t *TimeSeries) Add(v XY) {
	*t = append(*t, v)
}
