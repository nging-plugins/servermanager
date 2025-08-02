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
