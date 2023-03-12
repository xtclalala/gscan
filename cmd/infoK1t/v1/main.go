package main

import (
	"github.com/xtclalala/infoK1t/internal/output"
	"github.com/xtclalala/ylog"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"
)

var (
	target  string
	targets []string
)

func main() {
	SetCpuWorkerNum()

	var flags = []cli.Flag{
		&cli.UintFlag{
			Name:     "logLevel",
			Aliases:  []string{"ll"},
			Value:    8,
			Usage:    "set log level",
			Required: false,
		},
		&cli.BoolFlag{
			Name:     "color",
			Aliases:  []string{"c"},
			Value:    true,
			Usage:    "usage log color",
			Required: false,
		},
		&cli.UintFlag{
			Name:     "logMode",
			Aliases:  []string{"lm"},
			Usage:    "output log mode",
			Required: false,
			Value:    1,
		},
	}
	app := &cli.App{
		Name:    "infoK1t",
		Version: "0.0.1",
		Usage:   "usage",
		Flags:   flags,
		Commands: []*cli.Command{
			identify,
			probe,
			ping,
		},
		Before: func(c *cli.Context) error {
			var err error
			output.SetUseColor(c.Bool("color"))
			err = output.SetLogLevel(ylog.LogLevel(c.Uint("logLevel")))
			output.SetColor()
			output.AddTime()
			if err != nil {
				ylog.WithField("command", "main").Errorf(err.Error())
				return err
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		ylog.WithField("command", "main").Fatalf(err.Error())
	}
}

func SetCpuWorkerNum() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
