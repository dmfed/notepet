package main

import (
	"flag"
	"fmt"

	"github.com/dmfed/notepet"
)

func openStorage(storagename, storagetype string) (notepet.Storage, error) {
	var st notepet.Storage
	var err error
	switch storagetype {
	case "json":
		st, err = notepet.OpenOrInitJSONFileStorage(storagename)
	case "sqlite":
		st, err = notepet.OpenOrInitSQLiteStorage(storagename)
	case "network":
		st, err = nil, fmt.Errorf("Not implemented")
	default:
		st, err = nil, fmt.Errorf("Invalid argument passed")
	}
	return st, err
}

func main() {
	var (
		flagSource          = flag.String("src", "", "source storage")
		flagDestination     = flag.String("dst", "", "destination storage")
		flagSourceType      = flag.String("st", "", "type of source storage (json, sqlite, network)")
		flagDestinationType = flag.String("dt", "", "type of destination storage (json, sqlite, network)")
	)
	flag.Parse()
	src, err := openStorage(*flagSource, *flagSourceType)
	if err != nil {
		fmt.Println("failed to open source storage:", err)
	}
	dst, err := openStorage(*flagDestination, *flagDestinationType)
	if err != nil {
		fmt.Println("falied to open destination storage:", err)
	}
	if err := notepet.Migrate(dst, src); err != nil {
		fmt.Println("failed to migrate notes:", err)
	} else {
		fmt.Println("all done")
	}
}
