package main

import (
	"log"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/client"
)

var CLI struct {
	Start struct {
		Name string `arg:"" name:"name" help:"Name of the activity to start."`
	} `cmd:"" help:"Start an activity."`

	End struct {
	} `cmd:"" help:"End the current activity."`
}

func main() {

	conf, err := narc.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	cl := client.New(conf.ServerBaseURL, makeDaemon)
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "start <name>":
		{
			err = cl.StartActivity(CLI.Start.Name)
			if err != nil {
				ctx.Errorf("error starting activity: %s", err)
			}
		}
	case "end":
		{
			err = cl.StopActivity()
			if err != nil {
				ctx.Errorf("error starting activity: %s", err)
			}
		}
	default:
		panic(ctx.Command())
	}
}

func makeDaemon() error {
	cmd := exec.Command("go", "run", "/Users/alexander.jones/code/external/narc/cmd/daemon")
	return cmd.Start()
}
