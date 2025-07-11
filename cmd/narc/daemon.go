package main

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/daemon"
	"github.com/alexrjones/narc/idle"
	"github.com/alexrjones/narc/store"
)

type daemonOptions struct {
	initialActivityName string
	initialIgnoreIdle   bool
}

type daemonOption func(*daemonOptions)

type FuncYielder func() FuncYielder

func withInitialActivity(name string, ignoreIdle bool) daemonOption {
	return func(options *daemonOptions) {
		options.initialActivityName = name
		options.initialIgnoreIdle = ignoreIdle
	}
}

func daemonMain(c *narc.Config, logToFile bool, opts ...daemonOption) (invokeNext FuncYielder) {
	options := &daemonOptions{}
	for _, o := range opts {
		o(options)
	}
	if logToFile {
		cleanup, err := configureFileLogging(c.LogPath)
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()
	}
	var s daemon.Store
	if c.StorageType == narc.StorageTypeCSV {
		f, err := os.OpenFile(c.CSVPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		s = store.NewCSVStore(f)
	} else {
		panic(fmt.Sprintf("Unknown storage type %s", c.StorageType))
	}
	ctx, cancel := context.WithCancel(context.Background())
	m := idle.NewMonitor(c.IdleTimeout)
	d, err := daemon.New(s, m.Start(ctx))
	if err != nil {
		log.Fatal(err)
	}
	if options.initialActivityName != "" {
		err = d.SetActivity(ctx, options.initialActivityName, daemon.WithIgnoreIdle(options.initialIgnoreIdle))
		if err != nil {
			log.Fatal(err)
		}
	}
	d.Run(ctx)

	port := cmp.Or(c.ServerBaseURL[strings.LastIndex(c.ServerBaseURL, ":")+1:], "80")
	termChannel := make(chan daemon.SignalPacket, 1)
	serv := daemon.NewServer(d, s, termChannel)
	httpServer := &http.Server{Addr: "0.0.0.0:" + port, Handler: serv.GetHandler()}
	go func() {
		httpServer.ListenAndServe()
	}()
	log.Println("Server ready")
	// Wait for interrupt signal to gracefully shutdown the server with a timeout
	sig := <-termChannel

	// The context is used to inform the server it has N seconds to finish the request it is currently handling
	ctx, srvCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer srvCancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	if closer, ok := s.(io.Closer); ok {
		closer.Close()
	}
	cancel()

	if sig.Signal == daemon.SignalHup {
		log.Println("Received config reload signal")
		c, err = narc.GetConfig()
		if err != nil {
			log.Fatal(err)
		}
		var opts []daemonOption
		if sig.LastActivityName != "" {
			opts = append(opts, withInitialActivity(sig.LastActivityName, sig.LastActivityIgnoreIdle))
		}
		return func() FuncYielder {
			return daemonMain(c, logToFile, opts...)
		}
	}
	return nil
}

func configureFileLogging(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	log.SetOutput(f)
	ret := func() {
		log.SetOutput(os.Stderr)
		f.Close()
	}

	return ret, nil
}
