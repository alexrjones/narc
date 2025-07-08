package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/client"
)

var CLI struct {
	Start struct {
		Name []string `arg:"" name:"nameparts" help:"Name of the activity to start."`
	} `cmd:"" help:"Start an activity."`

	End struct {
	} `cmd:"" help:"End the current activity."`

	Daemon struct{} `cmd:"" help:"Start the daemon."`

	Terminate struct {
	} `cmd:"" help:"Terminate the daemon."`
}

func main() {

	conf, err := narc.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "start <nameparts>":
		{
			err = client.New(conf.ServerBaseURL, makeDaemon).StartActivity(strings.Join(CLI.Start.Name, " "))
			if err != nil {
				ctx.Errorf("error starting activity: %s", err)
			}
		}
	case "end":
		{
			err = client.New(conf.ServerBaseURL, makeDaemon).StopActivity()
			if err != nil {
				ctx.Errorf("error starting activity: %s", err)
			}
		}
	case "daemon":
		{
			daemonMain(conf)
		}
	case "terminate":
		{
			err = client.New(conf.ServerBaseURL, makeDaemon).TerminateDaemon()
			if err != nil {
				ctx.Errorf("error terminating daemon: %s", err)
			}
		}
	default:
		panic(ctx.Command())
	}
}

var makeDaemon = exec.Command(os.Args[0], "daemon").Start
