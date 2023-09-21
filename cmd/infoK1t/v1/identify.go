package main

import (
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/internal/output"
	"github.com/xtclalala/infoK1t/pkg"
	"github.com/xtclalala/infoK1t/pkg/device"
	"go.uber.org/zap"
)

// 获取当前设备网络信息
var identify = &cli.Command{
	Name:    "identify",
	Aliases: []string{"id"},
	Usage:   "Gets the current device ident information",
	Action: func(c *cli.Context) error {
		var (
			device *device.Device
			err    error
		)
		device, err = pkg.Identify()
		if err != nil {
			output.GetLogger().Error(device.Err.Error())
			return nil
		}
		output.GetLogger().Info("network info",
			zap.String("Pcap Name", device.PcapName),
			zap.String("Pcap Description", device.PcapDescription),
			zap.String("IPv4", device.Ipv4.String()),
			zap.String("NetMask", device.IpMask.String()),
			zap.String("Mac", device.Mac.String()),
		)
		return nil
	},
}
