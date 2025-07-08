package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/daemon"
	"github.com/alexrjones/narc/store"
)

func main() {

	c, err := narc.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	var s daemon.Store
	if c.StorageType == narc.StorageTypeCSV {
		f, err := os.Open(c.CSVPath)
		if err != nil {
			log.Fatal(err)
		}
		s = store.NewCSVStore(f)
	} else {
		panic(fmt.Sprintf("Unknown storage type %s", c.StorageType))
	}
	d, err := daemon.New(context.Background(), s)
	if err != nil {
		log.Fatal(err)
	}
	d.Run(context.Background())
	for {
		select {}
	}
}
