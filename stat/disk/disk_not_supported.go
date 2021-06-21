//+build !linux,!darwin

package disk

// Usage contains the disk stat details
type Disk struct {}

//GetDiskStatFromDir gets the Usage stat for the given dir
func GetDiskStatFromDir(dir string) (*Disk, error) {
	panic("system not supported")
}