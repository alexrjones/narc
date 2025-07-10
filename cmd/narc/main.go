package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alexrjones/narc"
	"github.com/alexrjones/narc/client"
)

var CLI struct {
	Start struct {
		Meeting bool `short:"m" help:"Runs in 'meeting mode': won't treat idle events as the end of an activity."`

		Name []string `arg:"" name:"nameparts" help:"Name of the activity to start."`
	} `cmd:"" help:"Start an activity."`

	End struct {
	} `cmd:"" help:"End the current activity."`

	Status struct {
	} `cmd:"" help:"Get the current status of the daemon and activity."`

	Aggregate struct {
		Round bool `default:"true" help:"Round durations to the nearest 5 minutes."`

		Start string `arg:"" optional:"" name:"start" help:"Start of the period over which to aggregate. Use time.DateOnly format or 'yesterday', 'today', 'tomorrow'."`
		End   string `arg:"" optional:"" name:"end" help:"End of the period over which to aggregate. Use time.DateOnly format or 'yesterday', 'today', 'tomorrow'."`
	} `cmd:"" aliases:"agg" help:"Aggregate time logs over the specified period."`

	Daemon struct{} `cmd:"" help:"Start the daemon."`

	Terminate struct {
	} `cmd:"" help:"Terminate the daemon."`

	Config struct {
		Show struct {
		} `cmd:"" help:"Prints current configuration."`
		Get struct {
			Name string `arg:"" name:"name" help:"Name of the config option."`
		} `cmd:"" help:"Print the value of a config option."`
		Set struct {
			Name  string `arg:"" name:"name" help:"Name of the config option."`
			Value string `arg:"" name:"value" help:"Value of the config option. The special value \"default\" will reset it to its default."`
		} `cmd:"" help:"Set a config option."`
	} `cmd:""`
}

func main() {

	conf, err := narc.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	makeDaemon := getMakeDaemon(conf)
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "start <nameparts>":
		{
			err = client.New(conf.ServerBaseURL, makeDaemon).StartActivity(strings.Join(CLI.Start.Name, " "), CLI.Start.Meeting)
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
	case "status":
		{
			res, err := client.New(conf.ServerBaseURL, makeDaemon).GetStatus()
			if err != nil {
				ctx.Errorf("error getting status: %s", err)
			} else {
				ctx.Printf(res)
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
	case "aggregate", "aggregate <start>", "aggregate <start> <end>":
		{
			start, err := parseTimeString(CLI.Aggregate.Start)
			if err != nil {
				ctx.Fatalf("error parsing start time: %s", err)
			}
			end, err := parseTimeString(CLI.Aggregate.End)
			if err != nil {
				ctx.Fatalf("error parsing end time: %s", err)
			}
			var agg string
			agg, err = client.New(conf.ServerBaseURL, makeDaemon).Aggregate(start, end, CLI.Aggregate.Round)
			if err != nil {
				ctx.Fatalf("error getting aggregate: %s", err)
			}
			fmt.Print(agg)
		}
	case "config show":
		{
			ctx.Printf("%s", conf)
		}
	case "config get <name>":
		{
			fmt.Println(conf.PropertyByName(CLI.Config.Get.Name))
		}
	case "config set <name> <value>":
		{
			err = narc.SetConfigOption(CLI.Config.Set.Name, CLI.Config.Set.Value)
			if err != nil {
				ctx.Errorf("failed to update config: %s", err)
			}
		}
	default:
		panic(ctx.Command())
	}
}

func parseTimeString(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	now := time.Now()
	s = strings.ToLower(s)
	if s == "today" {
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	} else if s == "tomorrow" {
		now = now.Add(time.Hour * 24)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	} else if s == "yesterday" {
		now = now.Add(time.Hour * -24)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	return time.Parse(time.DateOnly, s)
}

func getMakeDaemon(c *narc.Config) func() error {
	return func() error {
		f, err := os.OpenFile(c.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd := exec.Command(os.Args[0], "daemon")
		cmd.Stdout, cmd.Stderr = f, f
		return cmd.Start()
	}
}
