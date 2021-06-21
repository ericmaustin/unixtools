package zfs

import (
	"bytes"
	"encoding/json"
	cap "github.com/ericmaustin/unixtools/capacity"
	"github.com/ghodss/yaml"
	"os/exec"
	"strings"
)


type StateValue string

const (
	StateOnline   StateValue = "ONLINE"
	StateDegraded StateValue = "DEGRADED"
	StateFaulted  StateValue = "FAULTED"
	StateOffline  StateValue = "OFFLINE"
	StateUnavail  StateValue = "UNAVAILABLE"
	StateRemoved  StateValue = "REMOVED"
	StateAvail    StateValue = "IN USE"
	StateInUse    StateValue = "AVAILABLE"
)

func parseState(state string) StateValue {
	state = strings.ToLower(strings.TrimSpace(state))
	switch state {
	case "online":
		return StateOnline
	case "degraded":
		return StateDegraded
	case "faulted":
		return StateFaulted
	case "offline":
		return StateOffline
	case "removed":
		return StateRemoved
	case "inuse":
		return StateInUse
	case "avail":
		return StateAvail
	default:
		return StateUnavail
	}
}

type DevType string

const (
	DevTypeBlock  DevType = "block"
	DevTypeMirror DevType = "mirror"
	DevTypeRaidz1 DevType = "raidz1"
	DevTypeRaidz2 DevType = "raidz2"
	DevTypeRaidz3 DevType = "raidz3"
	DevTypeRaidz4 DevType = "raidz4"
	DevTypeSpare  DevType = "spare"
)

// parseDevType parses the device type from a string
func parseDevType(devType string) DevType {
	devType = strings.ToLower(strings.TrimSpace(devType))

	switch {
	case strings.HasPrefix(devType, "mirror"):
		return DevTypeMirror
	case strings.HasPrefix(devType, "spare"):
		return DevTypeSpare
	case strings.HasPrefix(devType, "raidz1"):
		return DevTypeRaidz1
	case strings.HasPrefix(devType, "raidz2"):
		return DevTypeRaidz2
	case strings.HasPrefix(devType, "raidz3"):
		return DevTypeRaidz3
	case strings.HasPrefix(devType, "raidz4"):
		return DevTypeRaidz4
	default:
		return DevTypeBlock
	}
}

// state represents the state of a block device, vdev or pool
type state struct {
	State    StateValue
	Read     uint64
	Write    uint64
	CheckSum uint64
	Message  string
}

// Zpool represents the status of a pool
type Zpool struct {
	Name                 string
	Size                 cap.Capacity `yaml:",omitempty" json:",omitempty"`
	Allocated            cap.Capacity `yaml:",omitempty" json:",omitempty"`
	Free                 cap.Capacity `yaml:",omitempty" json:",omitempty"`
	FragmentationPercent float64      `yaml:"fragmentation_percent,omitempty" json:"fragmentation_percent,omitempty"`
	CapacityPercent      float64      `yaml:"capacity_percent,omitempty" json:"capacity_percent,omitempty"`
	DeduplicationRatio   float64      `yaml:"deduplication_ratio,omitempty" json:"deduplication_ratio,omitempty"`
	Health               string       `yaml:",omitempty" json:",omitempty"`
	AltRoot              string       `yaml:",omitempty" json:",omitempty"`
	State                string
	Status               string      `yaml:",omitempty" json:",omitempty"`
	Action               string      `yaml:",omitempty" json:",omitempty"`
	See                  string      `yaml:",omitempty" json:",omitempty"`
	Scrub                string      `yaml:",omitempty" json:",omitempty"`
	Error                string      `yaml:",omitempty" json:",omitempty"`
	ReadErrors           uint64      `yaml:"read_errors,omitempty" json:"read_errors,omitempty"`
	WriteErrors          uint64      `yaml:"write_errors,omitempty" json:"write_errors,omitempty"`
	CheckSumErrors       uint64      `yaml:"check_sum_errors,omitempty" json:"check_sum_errors,omitempty"`
	Message              string      `yaml:"message,omitempty" json:"message,omitempty"`
	Devs                 []*DevState `yaml:"devices,omitempty" json:"devices,omitempty"`
	Spares               []*DevState `yaml:"spares,omitempty"  json:"spares,omitempty"`
}

// String implements stringer
func (z *Zpool) String() string {
	b, err := yaml.Marshal(z)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// JSONString prints the JSON value for this Zpool
func (z *Zpool) JSONString() string {
	b, err := json.Marshal(z)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// DevState represents the state of a device
type DevState struct {
	Name           string
	Type           DevType
	State          StateValue
	ReadErrors     uint64      `yaml:"read_errors,omitempty" json:"read_errors,omitempty"`
	WriteErrors    uint64      `yaml:"write_errors,omitempty" json:"write_errors,omitempty"`
	CheckSumErrors uint64      `yaml:"check_sum_errors,omitempty" json:"check_sum_errors,omitempty"`
	Message        string      `yaml:"message_errors,omitempty" json:"message_errors,omitempty"`
	Parent         *DevState   `yaml:"-" json:"-"`
	Children       []*DevState `yaml:"devices,omitempty" json:"devices,omitempty"`
}

// Zpools represents a list of Zpool
type Zpools []*Zpool

// String implements the Stringer interface
func (z *Zpools) String() string {
	b, err := yaml.Marshal(z)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// ZpoolList represents a list of PoolListRows
type ZpoolList []*ZpoolListRow

// String implements the Stringer interface
func (p *ZpoolList) String() string {
	b, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// GetPoolStatus gets a complete Zpools slice with all complete pool statuses
func (pl ZpoolList) GetPoolStatus() (Zpools, error) {
	var err error

	out := make([]*Zpool, len(pl))

	for i, row := range pl {
		if out[i], err = row.GetPoolStatus(); err != nil {
			return nil, err
		}
	}

	return out, nil
}

// ZpoolListRow represents the output from a zpool list command
type ZpoolListRow struct {
	Name                 string
	Size                 cap.Capacity `yaml:",omitempty" json:",omitempty"`
	Allocated            cap.Capacity `yaml:",omitempty" json:",omitempty"`
	Free                 cap.Capacity `yaml:",omitempty" json:",omitempty"`
	FragmentationPercent float64      `yaml:"fragmentation_percent,omitempty" json:"fragmentation_percent,omitempty"`
	CapacityPercent      float64      `yaml:"capacity_percent,omitempty" json:"capacity_percent,omitempty"`
	DeduplicationRatio   float64      `yaml:"deduplication_ratio,omitempty" json:"deduplication_ratio,omitempty"`
	Health               string       `yaml:",omitempty" json:",omitempty"`
	AltRoot              string       `yaml:",omitempty" json:",omitempty"`
}

// String implements the Stringer interface
func (z *ZpoolListRow) String() string {
	b, err := yaml.Marshal(z)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// GetPoolStatus gets the complete pool status
func (z *ZpoolListRow) GetPoolStatus() (*Zpool, error) {
	status, err := GetZpoolStatus(z.Name)
	if err != nil {
		return nil, err
	}

	status.Size = z.Size
	status.Allocated = z.Allocated
	status.Free = z.Free
	status.FragmentationPercent = z.FragmentationPercent
	status.CapacityPercent = z.CapacityPercent
	status.DeduplicationRatio = z.DeduplicationRatio
	status.Health = z.Health
	status.AltRoot = z.AltRoot

	return status, nil
}

// GetZpoolList runs a zpool list command and returns a ZpoolList
func GetZpoolList() (ZpoolList, error) {
	// zpool list -H -p -o name,size,allocated,free,fragmentation,capacity,dedupratio,health,altroot
	cmd := exec.Command("zpool", "list", "-H", "-p", "-o",
		"name,size,allocated,free,fragmentation,capacity,dedupratio,health,altroot")

	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil && out.Len() == 0 {
		return nil, err
	}

	return parseZpoolList(out.String()), nil
}

// GetZpoolStatus gets the given status for a given zpool name
func GetZpoolStatus(name string) (*Zpool, error) {
	cmd := exec.Command("zpool", "status", name)

	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil && out.Len() == 0 {
		return nil, err
	}

	return ParseZpoolStatus(out.String()), nil
}

