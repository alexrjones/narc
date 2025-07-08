package main

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/daemon"
	"github.com/alexrjones/narc/server"
	"github.com/alexrjones/narc/store"
)

func main() {

	c, err := narc.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	var s daemon.Store
	if c.StorageType == narc.StorageTypeCSV {
		f, err := os.OpenFile(c.CSVPath, os.O_RDWR|os.O_CREATE, 0644)
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

	port := cmp.Or(c.ServerBaseURL[strings.LastIndex(c.ServerBaseURL, ":")+1:], "8080")
	channel := make(chan struct{}, 1)
	serv := server.New(d, channel)
	httpServer := &http.Server{Addr: "0.0.0.0:" + port, Handler: serv.GetHandler()}
	go func() {
		httpServer.ListenAndServe()
	}()
	// Wait for interrupt signal to gracefully shutdown the server with a timeout
	<-channel

	// The context is used to inform the server it has N seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
