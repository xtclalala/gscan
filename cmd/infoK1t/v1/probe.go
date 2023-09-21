package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/internal/output"
	"github.com/xtclalala/infoK1t/pkg"
	"go.uber.org/zap"
	"time"
)

var probe = &cli.Command{
	Name:  "probe",
	Usage: "Probe all device information for the network where the current netmask",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "target",
			Aliases: []string{"t"},
			Usage:   "targets ip and netmask, use ',' split, explame:  22.22.22.0/24 or 11.11.11.0/24,33.33.0.0/16",
			Value:   "",
		},
	},
	Action: func(c *cli.Context) error {
		r := pkg.DefaultRunner()
		arpInfos, err := pkg.Probe(c.String("target"))
		if err != nil {
			output.GetLogger().Error(err.Error())
			r.DoneCh()
			r.Close()
			return nil
		}
		for {
			select {
			case data := <-arpInfos:
				output.GetLogger().Info(fmt.Sprintf("%s is done!", data.Ip),
					zap.String("IP", data.Ip),
					zap.String("Mac", data.Mac),
					zap.String("Device", data.M),
				)
			case <-time.After(10 * time.Second):
				output.GetLogger().Info("timeout")
				r.DoneCh()
				r.Close()
				return nil
			}
		}

	},
}
