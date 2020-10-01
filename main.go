package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/imunhatep/gorunner/system"
)

func main() {
	app := &cli.App{
		Name:      "gorunner",
		Usage:     "Parallel sub-process runner with graceful stop if any sub-process exits",
		UsageText: "gorunner --run='/bin/echo \"test1\"' --run='/bin/echo \"test2\"' -vv",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "run",
				Aliases:  []string{"r"},
				Usage:    "cli command to run, can be defined multiple times",
				Required: true,
			},
			&cli.DurationFlag{
				Name:     "timeout",
				Aliases:  []string{"t"},
				Value:    60 * time.Second,
				Usage:    "how long to wait before forcing others tasks to exit with sig kill",
				Required: false,
			},
			&cli.IntFlag{
				Name:     "maxprocs",
				Aliases:  []string{"j"},
				Value:    2,
				Usage:    "controls the number of operating system threads allocated to goroutines",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "v",
				Usage:    "log verbosity: info",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "vv",
				Usage:    "log verbosity: debug",
				Required: false,
			},
		},
		Action: func(args *cli.Context) error {
			runtime.GOMAXPROCS(args.Int("maxprocs"))

			setLogLevel(args)

			taskList, err := buildRunnableServiceList(args.StringSlice("run"))
			if err != nil {
				log.Fatal().Err(err).Msg(err.Error())
				return err
			}

			serviceMng := system.NewServiceManager(taskList, args.Duration("timeout"))

			return serviceMng.Run()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error().Err(err)
	}
}

func buildRunnableServiceList(commands []string) ([]*system.Service, error) {
	var tasks []*system.Service
	for i, command := range commands {
		cmdParts := strings.Split(command, " ")

		task := &system.Service{
			Name:    fmt.Sprintf("run%d", i),
			Command: cmdParts[0],
			Args:    cmdParts[1:],
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func setLogLevel(args *cli.Context) {
	switch true {
	case args.IsSet("vv"):
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case args.IsSet("v"):
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	fmt.Printf("Logging level: %s \n", zerolog.GlobalLevel().String())
}
