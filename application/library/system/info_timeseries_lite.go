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

import "sync"

var poolRealTimeStatusLite = sync.Pool{
	New: func() any {
		return &RealTimeStatusLite{pooled: true}
	},
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

func (r *RealTimeStatusLite) CopyTruncated(f *RealTimeStatus, max int) *RealTimeStatusLite {
	r.CPU = f.CPU.GetTruncate(max)
	r.Mem = f.Mem.GetTruncate(max)
	r.Net = NetIOTimeSeries{
		BytesSent:   f.Net.BytesSent.GetTruncate(max),
		BytesRecv:   f.Net.BytesRecv.GetTruncate(max),
		PacketsSent: f.Net.PacketsSent.GetTruncate(max),
		PacketsRecv: f.Net.PacketsRecv.GetTruncate(max),
	}
	r.Temp = make(map[string]TimeSeries, len(f.Temp))
	for key, value := range f.Temp {
		r.Temp[key] = value.GetTruncate(max)
	}
	return r
}
