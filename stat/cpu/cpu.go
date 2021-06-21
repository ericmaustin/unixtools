//+build linux,darwin

package cpu

import (
	"github.com/mackerelio/go-osstat/cpu"
	"time"
)

type CPU struct {
	FirstCount *cpu.Stats
	SecondCount *cpu.Stats
	Total float64
	User float64
	System float64
	Idle float64
}

func GetCPUStats(sleepDir time.Duration) (c *CPU, err error) {
	c = new(CPU)
	c.FirstCount, err = cpu.Get()
	if err != nil {
		return
	}
	time.Sleep(sleepDir)
	c.SecondCount, err = cpu.Get()
	if err != nil {
		return
	}
	c.Total = float64(c.SecondCount.Total - c.FirstCount.Total)
	c.User = float64(c.SecondCount.User-c.FirstCount.User)/c.Total*100
	c.System = float64(c.SecondCount.System-c.FirstCount.System)/c.Total*100
	c.Idle = float64(c.SecondCount.Idle-c.FirstCount.Idle)/c.Total*100
	return c, nil
}