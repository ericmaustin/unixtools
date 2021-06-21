//+build linux,darwin

package memory

import (
	"github.com/ericmaustin/unixtools/capacity"
	"github.com/mackerelio/go-osstat/memory"
)

// Mem represents memory stats
// wrapper around github.com/mackerelio/go-osstat/memory that returns the mem
// as bytesize.BS structs instead of uint64 values
type Mem struct {
	Total, Used, Cached, Free, Active, Inactive, SwapTotal, SwapUsed, SwapFree capacity.Capacity
}

//GetMemStats gets a Mem ptr with sizes in Capacity
func GetMemStats() *Mem {
	mem, err := memory.Get()
	if err != nil {
		panic(err)
	}
	return &Mem{
		Total:     capacity.Capacity(mem.Total),
		Used:      capacity.Capacity(mem.Used),
		Cached:    capacity.Capacity(mem.Cached),
		Free:      capacity.Capacity(mem.Free),
		Active:    capacity.Capacity(mem.Active),
		Inactive:  capacity.Capacity(mem.Inactive),
		SwapTotal: capacity.Capacity(mem.SwapTotal),
		SwapUsed:  capacity.Capacity(mem.SwapUsed),
		SwapFree:  capacity.Capacity(mem.SwapFree),
	}
}