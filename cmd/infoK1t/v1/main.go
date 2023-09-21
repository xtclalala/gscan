package main

import (
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/internal/config"
	"github.com/xtclalala/infoK1t/internal/output"
	"go.uber.org/zap"
	"os"
	"runtime"
)

func main() {
	SetCpuWorkerNum()

	var flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "mode",
			Aliases:  []string{"m"},
			Usage:    "output log mode",
			Required: false,
			Value:    "console",
		},
		&cli.StringFlag{
			Name:     "file",
			Aliases:  []string{"f"},
			Usage:    "output file path",
			Required: false,
			Value:    "",
		},
		&cli.StringFlag{
			Name:     "Gateway IP",
			Aliases:  []string{"g"},
			Usage:    "current network gateway ip",
			Required: false,
			Value:    "",
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
			subDomain,
		},
		Before: func(c *cli.Context) error {
			// init logger
			output.InitLogger(
				func() []string {
					outs := []string{"stdout"}
					if f := c.String("file"); f != "" {
						outs = append(outs, f)
					}
					outs = append(outs, config.GetOptions().Logger.File...)
					return outs
				}, func() string {
					if mode := c.String("mode"); mode != "" {
						return mode
					}
					return config.GetOptions().Logger.Mode
				}, func() (level zap.AtomicLevel, dev bool) {
					dev = config.IsDev()
					if dev {
						level = zap.NewAtomicLevelAt(zap.DebugLevel)
						return
					}
					level = zap.NewAtomicLevelAt(zap.InfoLevel)
					return
				})
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		output.GetLogger().Fatal(err.Error())
	}
}

func SetCpuWorkerNum() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
