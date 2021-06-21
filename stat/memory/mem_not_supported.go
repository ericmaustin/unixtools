//+build !linux,!darwin

package memory

// Mem represents memory stats
// wrapper around github.com/mackerelio/go-osstat/memory that returns the mem
// as bytesize.BS structs instead of uint64 values
type Mem struct {}

//GetMemStats gets a Mem ptr with sizes in Capacity
func GetMemStats() *Mem {
	panic("system not supported")
}