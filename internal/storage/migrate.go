package storage

import (
	"flag"
	"fmt"

	"github.com/dmfed/notepet"
)

// Migrate copies all model.Notes from src (source) Storage
// to dst (destination) Storage. If succesful the returned
// error in nil.
func Migrate(dst, src Storage) error {
	if src == nil || dst == nil {
		return ErrStorageIsNil
	}
	notes, err := src.Get()
	if err != nil {
		return err
	}
	for i := len(notes) - 1; i >= 0; i-- {
		if _, err := dst.Put(notes[i]); err != nil {
			return err
		}
	}
	return nil
}

func openStorage(storagename, storagetype string) (notepet.Storage, error) {
	var st notepet.Storage
	var err error
	switch storagetype {
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
		return
	}
	dst, err := openStorage(*flagDestination, *flagDestinationType)
	if err != nil {
		fmt.Println("failed to open destination storage:", err)
		return
	}
	if err := notepet.Migrate(dst, src); err != nil {
		fmt.Println("failed to migrate notes:", err)
	} else {
		fmt.Println("all done")
	}
}
