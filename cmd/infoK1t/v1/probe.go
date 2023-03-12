package main

import (
	"github.com/google/gopacket"
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/pkg"
	"github.com/xtclalala/infoK1t/pkg/device"
	"github.com/xtclalala/infoK1t/pkg/protocol"
	"github.com/xtclalala/infoK1t/pkg/yIP"
	"github.com/xtclalala/ylog"
	"net"
	"time"
)

var probe = &cli.Command{
	Name:  "probe",
	Usage: "Probe all device information for the network where the current netmask",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {
		var (
			arp *protocol.ArpProtocol
			d   *device.Device
		)

		d = pkg.DefaultDevice()
		arp = protocol.NewArpProtocol()

		var (
			srcIp     = d.Ipv4.To4()
			srcIpInt  int
			maskSize  int
			targetIps []int
			err       error
		)
		srcIpInt, err = yIP.Ip2int(srcIp.String())
		if err != nil {
			ylog.WithField("command", "probe").Errorf("srcIp 'string' to 'int' is failed: " + err.Error())
			return nil
		}
		maskSize, _ = d.IpMask.Size()
		ylog.WithField("command", "probe").Debugf(string(maskSize))
		targetIps, err = yIP.Parse(srcIpInt, maskSize)
		if err != nil {
			ylog.WithField("command", "probe").Errorf(err.Error())
			return nil
		}

		var (
			temps = make([]net.IP, 0, 255)
		)
		for _, ip := range targetIps {
			temps = append(temps, net.ParseIP(yIP.Int2Ip(ip)))
		}

		packetCh := arp.BuildSendPacket(srcIp, d.Mac, temps)

		r := pkg.DefaultRunner()
		if r.Err != nil {
			ylog.WithField("command", "probe").Errorf(err.Error())
			return nil
		}

		r.AppendParseHandle(func(packet gopacket.Packet) bool {
			return arp.Parse(packet)
		})

		go r.RunSender()
		go r.RunReceive()

		for packet := range packetCh {
			r.PushPacket(packet)
		}
		for {
			select {
			case data := <-arp.ArpTable:
				ylog.WithFields(map[string]string{
					"command": "probe",
					"IP":      data.Ip,
					"Mac":     data.Mac,
					"Device":  data.M}).Infof("%s is done!", data.Ip)
			case <-time.After(5 * time.Second):
				ylog.WithField("command", "probe").Infof("timeout")
				r.DoneCh()
				r.Close()
				return nil
			}
		}

	},
}
