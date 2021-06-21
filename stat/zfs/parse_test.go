package zfs

import (
	"fmt"
	"testing"
)

var sampleZfs = `
  pool: zeepool
 state: DEGRADED
status: One or more devices could not be opened.  Sufficient replicas exist for
        the pool to continue functioning in a degraded state.
action: Attach the missing device and online it using 'zpool online'.
   see: http://www.sun.com/msg/ZFS-8000-2Q
 scrub: none requested
config:

        NAME          STATE     READ WRITE CKSUM
        zeepool       ONLINE       0     0     0
          mirror-0    ONLINE       0     0     0
            c1t2d0    ONLINE       0     0     0
            spare-1   ONLINE       0     0     0
              c2t3d0  ONLINE       0     0     0  90K resilvered
              c2t1d0  ONLINE       0     0     0
        spares
          c2t3d0      INUSE     currently in use

errors: No known data errors
`

var sampleList = `
boot-pool       33822867456     1469919232      32352948224     0       4       1.00    ONLINE  -
tank    41970420416512  20504386473984  21466033942528  0       48      1.00    ONLINE  /mnt
`

func TestParseStatus(t *testing.T) {
	p := ParseZpoolStatus(sampleZfs)
	fmt.Println(p.String())

	//fmt.Println("state:", p.state)
	//fmt.Println("status:", p.Status)
	//fmt.Println("action:", p.Action)
	//fmt.Println("scrub:", p.Scrub)
	//fmt.Println("errors:", p.Error)
	//
	//for _, dev := range p.Devs {
	//	fmt.Println(dev.Name, dev.Type.String())
	//
	//	for _, _dev := range dev.Children {
	//		fmt.Println("  ", _dev.Name, _dev.Type.String(), _dev.state.state)
	//	}
	//
	//}
	//
	//for _, dev := range p.Spares {
	//	fmt.Println(dev.Name, dev.Type.String(), dev.state.Message)
	//}
}

func TestGetKey(t *testing.T) {
	fmt.Println(getFieldValue("status", sampleZfs, false))
	fmt.Println(getFieldValue("spares", sampleZfs, true))
}


func TestParseZpoolList(t *testing.T) {
	p := parseZpoolList(sampleList)
	fmt.Println(p.String())
}