package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/pkg"
	"github.com/xtclalala/infoK1t/pkg/device"
	"github.com/xtclalala/ylog"
)

// 获取当前设备网络信息
var identify = &cli.Command{
	Name:    "identify",
	Aliases: []string{"id"},
	Usage:   "Gets the current device ident information",
	Action: func(c *cli.Context) error {
		var d *device.Device
		d = pkg.DefaultDevice()

		if d.Err != nil {
			ylog.WithField("command", "identify").Errorf("device collect info is failed")
			ylog.WithField("command", "identify").Debugf(d.Err.Error())
			return nil
		}

		ylog.WithFields(map[string]string{
			fmt.Sprintf("%-20s", "Pcap Name"):        d.PcapName,
			fmt.Sprintf("%-20s", "Pcap Description"): d.PcapDescription,
			fmt.Sprintf("%-20s", "IPv4"):             d.Ipv4.String(),
			fmt.Sprintf("%-20s", "NetMask"):          d.IpMask.String(),
			fmt.Sprintf("%-20s", "Mac"):              d.Mac.String(),
		}).Infof("Finish")
		return nil
	},
}
