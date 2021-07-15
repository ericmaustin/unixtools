//+build linux darwin

package disk

import (
	"fmt"
	cap "github.com/ericmaustin/unixtools/capacity"
	"github.com/jaypipes/ghw/pkg/block"
	"gopkg.in/yaml.v2"
	"strings"
	"sync"
	"syscall"
	"time"
)

// BlockDevice represents a storage device
type BlockDevice struct {
	*block.Disk
	DevID                  int32        `yaml:"dev_id,omitempty" json:"dev_id,omitempty"`
	Partitions             []*Partition `yaml:"partitions,omitempty" json:"partitions,omitempty"`
	SizeBytes              cap.Capacity `yaml:"size" json:"size"`
	PhysicalBlockSizeBytes cap.Capacity `yaml:"physical_block_size_bytes" json:"physical_block_size_bytes"`
	SMART                  *SMARTInfo
	mu                     sync.Mutex
	lastSmartCall          time.Time
}

// String implements stringer and returns a yaml formatted string
func (d *BlockDevice) String() string {
	b, err := yaml.Marshal(d)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// SMARTInfo gets smart info for the disk
func (d *BlockDevice) SMARTInfo() (*SMARTInfo, error) {
	var err error

	d.mu.Lock()

	if d.lastSmartCall.Add(time.Minute).Before(time.Now()) {
		d.SMART, err = GetSMARTInfo(d.Name)
	}

	d.mu.Unlock()

	return d.SMART, err
}

// Partition represents a logical block device partition
type Partition struct {
	Name       string
	Label      string
	MountPoint string
	Type       string
	IsReadOnly bool
	UUID       string // This would be volume UUID on macOS, PartUUID on linux, empty on Windows
	Disk       *BlockDevice
	SizeBytes  cap.Capacity
	Capacity   *FsCapacity
}

// PartitionInfo used for printing / marshaling to prevent recursive marshaling calls
type PartitionInfo struct {
	Name       string             `yaml:"name" json:"name"`
	Label      string             `yaml:"label" json:"label"`
	MountPoint string             `yaml:"mount_point" json:"mount_point"`
	Type       string             `yaml:"type" json:"type"`
	IsReadOnly bool               `yaml:"read_only" json:"read_only"`
	UUID       string             `yaml:"uuid" json:"uuid"` // This would be volume UUID on macOS, PartUUID on linux, empty on Windows
	Disk       *PartitionDiskInfo `yaml:"disk,omitempty" json:"disk,omitempty"`
	SizeBytes  cap.Capacity       `yaml:"size" json:"size"`
	Capacity   *FsCapacity        `yaml:"capacity,omitempty" json:"capacity,omitempty"`
}

// PartitionDiskInfo used for printing / marshaling to prevent recursive marshaling calls
type PartitionDiskInfo struct {
	DevID                  int32        `yaml:"dev_id,omitempty" json:"dev_id,omitempty"`
	SizeBytes              cap.Capacity `yaml:"size" json:"size"`
	PhysicalBlockSizeBytes cap.Capacity `yaml:"physical_block_size_bytes" json:"physical_block_size_bytes"`
	SMART                  *SMARTInfo   `yaml:"smart,omitempty" json:"smart,omitempty"`
}

func (p *Partition) toInfo() *PartitionInfo {
	return &PartitionInfo{
		Name:       p.Name,
		Label:      p.Label,
		MountPoint: p.MountPoint,
		Type:       p.Type,
		IsReadOnly: p.IsReadOnly,
		UUID:       p.UUID,
		Disk: &PartitionDiskInfo{
			DevID:                  p.Disk.DevID,
			SizeBytes:              p.Disk.SizeBytes,
			PhysicalBlockSizeBytes: p.Disk.PhysicalBlockSizeBytes,
			SMART:                  p.Disk.SMART,
		},
		SizeBytes: p.SizeBytes,
		Capacity:  p.Capacity,
	}
}

// String prints the yaml string
func (p *Partition) String() string {
	info := p.toInfo()
	b, err := yaml.Marshal(info)

	if err != nil {
		panic(err)
	}

	return string(b)
}

type FsCapacity struct {
	syscall.Statfs_t
	UsedBytes      cap.Capacity `yaml:"used_bytes" json:"used_bytes"`
	AvailableBytes cap.Capacity `yaml:"available" json:"available"`
	TotalBytes     cap.Capacity `yaml:"total" json:"total"`
	BlockSizeBytes cap.Capacity `yaml:"block_size" json:"block_size"`
}

// GetPartitionFromDir gets the filesystem capacity for the given dir
func GetPartitionFromDir(dir string) (*Partition, error) {
	var stat syscall.Stat_t

	if err := syscall.Stat(dir, &stat); err != nil {
		return nil, err
	}

	if fs, ok := partMap[stat.Dev]; ok {
		return fs, nil
	}

	return nil, fmt.Errorf("could not find filesytem for %s", dir)
}

// GetDiskFromLabel gets the disk from a device label string
func GetDiskFromLabel(dev string) (*BlockDevice, error) {
	dev = strings.TrimPrefix(dev, "/dev/")

	if d, ok := devMap[dev]; ok {
		return d, nil
	}

	return nil, fmt.Errorf("could not find disk with label %s", dev)
}

// GetFSCapacity gets the filesystem capacity for the given dir
func GetFSCapacity(dir string) (*FsCapacity, error) {
	ds := new(FsCapacity)
	if err := syscall.Statfs(dir, &ds.Statfs_t); err != nil {
		return nil, err
	}

	ds.BlockSizeBytes = cap.Capacity(ds.Statfs_t.Bsize)
	ds.AvailableBytes = ds.BlockSizeBytes.Mult(int64(ds.Statfs_t.Bavail))
	ds.TotalBytes = ds.BlockSizeBytes.Mult(int64(ds.Statfs_t.Blocks))
	ds.UsedBytes = ds.TotalBytes.Sub(ds.AvailableBytes)

	return ds, nil
}

var (
	blockDevices *block.Info
	partMap      = make(map[int32]*Partition)
	devMap       = make(map[string]*BlockDevice)
	fsMu         sync.Mutex
)

// LoadMountedFileSystems loads all the mounted filesystems
func LoadMountedFileSystems() error {
	var (
		err  error
		stat syscall.Stat_t
	)

	fsMu.Lock()
	defer fsMu.Unlock()

	blockDevices, err = block.New()
	if err != nil {
		return err
	}

	for _, d := range blockDevices.Disks {
		disk := &BlockDevice{Disk: d}
		disk.SizeBytes = cap.Capacity(disk.Disk.SizeBytes)
		disk.PhysicalBlockSizeBytes = cap.Capacity(disk.Disk.PhysicalBlockSizeBytes)
		devMap[d.Name] = disk

		disk.Partitions = make([]*Partition, len(d.Partitions))

		for i, p := range d.Partitions {
			part := &Partition{
				Name:       p.Name,
				Label:      p.Label,
				MountPoint: p.MountPoint,
				Type:       p.Type,
				IsReadOnly: p.IsReadOnly,
				UUID:       p.UUID,
				Disk:       disk,
			}
			disk.Partitions[i] = part
			part.SizeBytes = cap.Capacity(p.SizeBytes)

			if len(p.MountPoint) < 1 {
				// skip all unmounted filesystems
				continue
			}

			if err = syscall.Stat(p.MountPoint, &stat); err != nil {
				return err
			}

			disk.DevID = stat.Dev

			part.Capacity, err = GetFSCapacity(p.MountPoint)
			if err != nil {
				return err
			}

			partMap[disk.DevID] = part
		}
	}

	return nil
}

func init() {
	var err error
	if err = LoadMountedFileSystems(); err != nil {
		panic(err)
	}
}

func boolYesNo(b bool) string {
	if b {
		return "Yes"
	}

	return "No"
}
