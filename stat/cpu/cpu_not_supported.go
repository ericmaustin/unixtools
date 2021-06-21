//+build !linux,!darwin

package cpu

import (
	"time"
)

type CPU struct {}

func GetCPUStats(sleepDir time.Duration) (c *CPU, err error) {
	panic("system not supported")
}