package zfs

import (
	"bytes"
	cap "github.com/ericmaustin/unixtools/capacity"
	"regexp"
	"strconv"
	"strings"
)

var (
	configTitleRe = regexp.MustCompile(`\s+NAME\s+STATE(.*)`)
	configLineRe  = regexp.MustCompile(`([^\s]+)\s+([^\s]+)\s+(\d+)\s+(\d+)\s+(\d+)(.*)`)
	configSpareRe = regexp.MustCompile(`([^\s]+)\s+([^\s]+)(.*)`)
)

// ParseZpoolStatus parses the output of a zpool status command
func ParseZpoolStatus(v string) *Zpool {
	var p = &Zpool{
		Name:   strings.TrimSpace(getFieldValue("pool", v, false)),
		State:  strings.TrimSpace(getFieldValue("state", v, false)),
		Status: strings.TrimSpace(getFieldValue("status", v, false)),
		Action: strings.TrimSpace(getFieldValue("action", v, false)),
		Scrub:  strings.TrimSpace(getFieldValue("scrub", v, false)),
		See:    strings.TrimSpace(getFieldValue("see", v, false)),
		Error:  strings.TrimSpace(getFieldValue("errors", v, false)),
	}

	config := getFieldValue("config", v, true)

	parseZpoolConfigLines(config, p)

	spares := getFieldValue("spares", v, false)

	if len(spares) > 0 {
		parseZpoolSpares(spares, p)
	}

	return p
}

func parseZpoolConfigLines(v string, poolStatus *Zpool) {
	lines := strings.Split(v, "\n")

	var (
		i, baseIndent, depth, prevDepth int
		currentDev                      *DevState
	)

	for _, line := range lines {
		if len(line) < 1 || configTitleRe.MatchString(line) {
			// skip empty and title lines
			continue
		}

		if strings.TrimSpace(line) == "spares" {
			// don't load spare data
			return
		}

		if i == 0 {
			_, state := parseStateLine(line)
			baseIndent = countIndent(line)
			poolStatus.ReadErrors = state.Read
			poolStatus.WriteErrors = state.Write
			poolStatus.CheckSumErrors = state.CheckSum
			poolStatus.Message = state.Message
			i++

			continue
		}

		// indent is the depth of the indentation
		depth = countIndent(line) - baseIndent
		if depth < 0 {
			// likely at the end of the config lines or unexpected output
			continue
		}

		dev := parseDevLine(line)

		switch true {
		case depth == 1:
			// root device
			poolStatus.Devs = append(poolStatus.Devs, dev)

		case depth > prevDepth:
			// indented
			dev.Parent = currentDev
			currentDev.Children = append(currentDev.Children, dev)

		case prevDepth == depth:
			dev.Parent = currentDev.Parent
			// same indent, add this dev to the parent's children
			currentDev.Parent.Children = append(currentDev.Parent.Children, dev)

		case prevDepth < prevDepth:
			// we've moved back up a dev, but we're not at the root
			// so we need to add this dev to the parent parent's children
			dev.Parent = currentDev.Parent.Parent
			currentDev.Parent.Parent.Children = append(currentDev.Parent.Parent.Children, dev)
		}

		currentDev = dev
		prevDepth = depth
		i++
	}
}

func parseZpoolSpares(v string, status *Zpool) {
	lines := strings.Split(v, "\n")

	for _, line := range lines {
		if len(line) < 1 {
			// skip empty lines
			continue
		}

		parts := configSpareRe.FindStringSubmatch(line)

		if len(parts) < 3 {
			// bad line?
			continue
		}

		dev := &DevState{
			Name:  parts[1],
			Type:  parseDevType(parts[1]),
			State: parseState(parts[2]),
		}

		if len(parts) == 4 {
			dev.Message = strings.TrimSpace(parts[3])
		}

		status.Spares = append(status.Spares, dev)
	}
}

func countIndent(in string) int {
	for i, ii := 0, 0; i < len(in); i, ii = i+2, ii+1 {
		if in[i:i+2] != "  " {
			return ii
		}
	}

	return -1
}

func parseDevLine(line string) *DevState {
	name, state := parseStateLine(line)

	return &DevState{
		Name:           name,
		Type:           parseDevType(name), //todo: parse type
		State:          state.State,
		ReadErrors:     state.Read,
		WriteErrors:    state.Write,
		CheckSumErrors: state.CheckSum,
		Message:        state.Message,
		Children:       []*DevState{},
	}
}

func getFieldValue(key string, v string, keepNewline bool) string {
	keyLen := len(key)
	idx := strings.Index(v, key) + keyLen

	var out, tmp bytes.Buffer

	for {
		idx++
		if v[idx] == '\n' {
			if keepNewline {
				out.WriteByte('\n')
			}

			out.Write(tmp.Bytes())

			if len(v) > idx+keyLen && len(strings.TrimSpace(v[idx:idx+keyLen+1])) < 1 {
				tmp.Reset()

				idx += keyLen + 1

				continue
			}

			return out.String()
		}

		tmp.WriteByte(v[idx])
	}
}

func parseStateLine(line string) (string, *state) {
	var (
		err   error
		parts = configLineRe.FindStringSubmatch(line)
	)

	if len(parts) < 6 {
		// invalid number of values
		return "", nil
	}

	s := &state{
		State: parseState(parts[2]),
	}

	if s.Read, err = strconv.ParseUint(parts[3], 10, 64); err != nil {
		panic(err)
	}

	if s.Write, err = strconv.ParseUint(parts[4], 10, 64); err != nil {
		panic(err)
	}

	if s.CheckSum, err = strconv.ParseUint(parts[5], 10, 64); err != nil {
		panic(err)
	}

	if len(parts) > 5 {
		s.Message = strings.TrimSpace(strings.Join(parts[6:], "; "))
	}

	return parts[1], s
}

// FlexSplit splits a string by a delim even if there are more than one delim value between fields
func FlexSplit(v string, delim string) []string {
	var (
		tmp bytes.Buffer
		out []string
	)

	delimLen := len(delim)
	vLen := len(v)

	for i := 0; i < vLen; i++ {
		if vLen >= i+delimLen && v[i:i+delimLen] == delim {
			if tmp.Len() > 0 {
				out = append(out, tmp.String())
				tmp.Reset()
			}

			i += delimLen - 1

			continue
		}

		tmp.WriteByte(v[i])
	}

	return out
}

func parseZpoolList(v string) ZpoolList {
	var out []*ZpoolListRow

	lines := strings.Split(v, "\n")

	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			// skip blank
			continue
		}

		fields := FlexSplit(l, " ")

		size, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			panic(err)
		}

		allocated, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			panic(err)
		}

		free, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			panic(err)
		}

		frag, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			panic(err)
		}

		dedup, err := strconv.ParseFloat(fields[5], 64)
		if err != nil {
			panic(err)
		}

		out = append(out, &ZpoolListRow{
			Name:                 fields[0],
			Size:                 cap.Capacity(size),
			Allocated:            cap.Capacity(allocated),
			Free:                 cap.Capacity(free),
			FragmentationPercent: frag,
			CapacityPercent:      dedup,
			DeduplicationRatio:   dedup,
			Health:               fields[6],
			AltRoot:              fields[7],
		})
	}

	return out
}
