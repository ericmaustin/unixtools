package capacity

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"strconv"
	"strings"
)

type units string

const (
	B   units = "B"
	KB        = "KB"
	KiB       = "KiB"
	MB        = "MB"
	MiB       = "MiB"
	GB        = "GB"
	GiB       = "GiB"
	TB        = "TB"
	TiB       = "TiB"
	PB        = "PB"
	PiB       = "PiB"
	EB        = "EB"
	EiB       = "EiB"
)

type measurementType string

func (u units) getMeasurementType() measurementType {
	if u == B {
		return Base10
	}

	if u[1] == 'i' {
		return Base2
	}

	return Base10
}

const (
	// Base2 implies measurement is based on IEC/SCI standard
	Base2 measurementType = "Base2"
	// Base10 implies measurement is based on modern base 10 measurements
	Base10 measurementType = "Base10"
)

// Capacity is used for byte sizes
type Capacity int64

// UnmarshalYAML unmarshals the yaml
func (cap *Capacity) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var (
		err    error
		target string
	)

	if err = unmarshal(&target); err != nil {
		return err
	}

	parsed, err := Parse(target)
	if err != nil {
		return err
	}

	*cap = *parsed

	return nil
}

// MarshalYAML implements the yaml Marshaler
// returns the string value of the capacity
func (cap Capacity) MarshalYAML() (interface{}, error) {
	str := cap.FormatBytes()
	return yaml.Marshal(str)
}

// B returns the the Capacity in bytes
func (cap Capacity) B() int64 {
	return int64(cap)
}

// KB returns the the Capacity in kilobytes
func (cap Capacity) KB() float64 {
	return float64(cap) / 1000
}

// KiB returns the the Capacity in kilobits
func (cap Capacity) KiB() float64 {
	return float64(cap) / 1024
}

// MB returns the the Capacity in megabytes
func (cap Capacity) MB() float64 {
	return float64(cap) / 1e6
}

// MiB returns the the Capacity in mebibytes
func (cap Capacity) MiB() float64 {
	return cap.KiB() / 1024
}

// GB returns the the Capacity in gigabytes
func (cap Capacity) GB() float64 {
	return float64(cap) / 1e9
}

// GiB returns the the Capacity in gibibytes
func (cap Capacity) GiB() float64 {
	return cap.MiB() / 1024
}

// TB returns the the Capacity in SCI terabytes
func (cap Capacity) TB() float64 {
	return float64(cap) / 1e12
}

// TiB returns the the Capacity in tebibyte
func (cap Capacity) TiB() float64 {
	return cap.GiB() / 1024
}

// PB returns the the Capacity in petabytes
func (cap Capacity) PB() float64 {
	return float64(cap) / 1e15
}

// PiB returns the the Capacity in pebibytes
func (cap Capacity) PiB() float64 {
	return cap.TiB() / 1024
}

// EB returns the the Capacity in exabytes
func (cap Capacity) EB() float64 {
	return float64(cap) / 1e18
}

// EiB returns the the Capacity in exbibytes
func (cap Capacity) EiB() float64 {
	return cap.PiB() / 1024
}

// Sub subtracts another Capacity
func (cap Capacity) Sub(other Capacity) Capacity {
	return Capacity(int64(cap) - int64(other))
}

// Add adds another Capacity
func (cap Capacity) Add(other Capacity) Capacity {
	return Capacity(int64(cap) + int64(other))
}

// Mult multiplies this Capacity by some int64 val and returns a new Capacity
func (cap Capacity) Mult(other int64) Capacity {
	return Capacity(int64(cap) * other)
}

// Div divides this Capacity by some int64 val and returns a new Capacity
func (cap Capacity) Div(other int64) Capacity {
	return Capacity(int64(cap) / other)
}

// String prints the Capacity as a smart-formatted string according to its size
func (cap Capacity) String() string {
	return cap.FormatBytes()
}

// FormatBytes prints the Capacity as a formatted string according to its size
func (cap Capacity) FormatBytes() string {
	v := int64(cap)

	var i int

	suffixes := [...]string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	d := int64(1)

	for i = 0; i < len(suffixes)-1; i++ {
		if v < d*1000 {
			break
		}

		d *= 1000
	}

	if i == 0 {
		return fmt.Sprintf("%d %s", v, suffixes[i])
	}

	return fmt.Sprintf("%.3g %s", float64(v)/float64(d), suffixes[i])
}

// FormatBase2Bytes prints the Capacity as a formatted string according to its size
// using IEC notation
func (cap Capacity) FormatBase2Bytes() string {
	v := int64(cap)

	var i int

	suffixes := [...]string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	d := int64(1)

	for i = 0; i < len(suffixes)-1; i++ {
		if v < d*1024 {
			break
		}

		d *= 1024
	}

	if i == 0 {
		return fmt.Sprintf("%d %s", v, suffixes[i])
	}

	return fmt.Sprintf("%.3g %s", float64(v)/float64(d), suffixes[i])
}

// Parse parses a string and returns a capacity
func Parse(v string) (*Capacity, error) {
	nv, unit, err := getValueAndUnits(v)
	if err != nil {
		return nil, err
	}

	return NewCapacity(nv, unit), nil
}

// ParseAs parses a string as the given measurement type and returns a capacity
func ParseAs(v string, mt measurementType) (*Capacity, error) {
	nv, unit, err := getValueAndUnits(v)
	if err != nil {
		return nil, err
	}

	if mt == Base2 {
		return parseAsBase2(nv, unit), nil
	}

	return parseAsBase10(nv, unit), nil
}

func getValueAndUnits(v string) (float64, units, error) {
	var (
		num, unitRaw string
		numDone      bool
		unit         units
	)

	v = strings.TrimSpace(v)

	if len(v) == 0 {
		return 0, "", nil
	}

	for i := 0; i < len(v); i++ {
		if (v[i] >= '0' && v[i] <= '9') || v[i] == '.' {
			if numDone {
				return 0, "", fmt.Errorf("%s is not a valid capacity", v)
			}

			num += string(v[i])

			continue
		}

		if v[i] == 0 {
			return 0, "", fmt.Errorf("%s is not a valid capacity", v)
		}

		numDone = true
		unitRaw += string(v[i])
	}

	switch strings.TrimSpace(strings.ToLower(unitRaw)) {
	case "", "b":
		unit = B
	case "k", "kb":
		unit = KB
	case "kib":
		unit = KiB
	case "m", "mb":
		unit = MB
	case "mib":
		unit = MiB
	case "g", "gb":
		unit = GB
	case "gib":
		unit = GiB
	case "t", "tb":
		unit = TB
	case "tib":
		unit = TiB
	case "p", "pb":
		unit = PB
	case "pib":
		unit = PiB
	case "e", "eb":
		unit = EB
	case "eib":
		unit = EiB
	default:
		return 0, "", fmt.Errorf("%s is not a valid unit type", unitRaw)
	}

	nv, err := strconv.ParseFloat(num, 64)

	if err != nil {
		return 0, "", err
	}

	return nv, unit, nil
}

func parseAsBase2(v float64, unit units) *Capacity {
	switch unit {
	case B:
		return NewCapacity(v, B)
	case KB, KiB:
		return NewCapacity(v, KiB)
	case MB, MiB:
		return NewCapacity(v, MiB)
	case GB, GiB:
		return NewCapacity(v, GiB)
	case TB, TiB:
		return NewCapacity(v, TiB)
	case PB, PiB:
		return NewCapacity(v, PiB)
	case EB, EiB:
		return NewCapacity(v, EiB)
	}

	panic(fmt.Errorf("%s is not a valid unit", unit))
}

func parseAsBase10(v float64, unit units) *Capacity {
	switch unit {
	case B:
		return NewCapacity(v, B)
	case KB, KiB:
		return NewCapacity(v, KB)
	case MB, MiB:
		return NewCapacity(v, MB)
	case GB, GiB:
		return NewCapacity(v, GB)
	case TB, TiB:
		return NewCapacity(v, TB)
	case PB, PiB:
		return NewCapacity(v, PB)
	case EB, EiB:
		return NewCapacity(v, EB)
	}

	panic(fmt.Errorf("%s is not a valid unit", unit))
}

// NewCapacity creates a new capacity with a given value and Unit of measurement
func NewCapacity(v float64, u units) *Capacity {
	var c Capacity

	switch u {
	case B:
		c = Capacity(v)
	case KB:
		c = Capacity(v * 1e3)
	case KiB:
		c = Capacity(v * 1024)
	case MB:
		c = Capacity(v * 1e6)
	case MiB:
		c = Capacity(v * (1 << 20))
	case GB:
		c = Capacity(v * 1e9)
	case GiB:
		c = Capacity(v * (1 << 30))
	case TB:
		c = Capacity(v * 1e12)
	case TiB:
		c = Capacity(v * (1 << 40))
	case PB:
		c = Capacity(v * 1e15)
	case PiB:
		c = Capacity(v * (1 << 50))
	case EB:
		c = Capacity(v * 1e18)
	case EiB:
		c = Capacity(v * (1 << 60))
	}

	return &c
}
