package main

import (
	"flag"
	"fmt"
	"github.com/ericmaustin/unixtools/stat/disk"
	"os"
	"runtime"
)

func main() {
	var err error

	fmt.Println("Go DISK info utility")
	fmt.Printf("Built with %s on %s (%s)\n\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	dir := flag.String("dir", "", "dir")

	if len(*dir) < 1 {
		dir = new(string)
		*dir, err = os.Getwd()

		if err != nil {
			panic(err)
		}
	}

	partition, err := disk.GetPartitionFromDir(*dir)
	if err != nil {
		panic(err)
	}

	fmt.Println(partition.String())
}

