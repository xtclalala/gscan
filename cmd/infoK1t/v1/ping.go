package main

import (
	"context"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/internal/timer"
	"github.com/xtclalala/infoK1t/pkg"
	"github.com/xtclalala/infoK1t/pkg/device"
	"github.com/xtclalala/infoK1t/pkg/protocol"
	"github.com/xtclalala/infoK1t/pkg/runner"
	"github.com/xtclalala/ylog"
	"net"
	"strings"
	"time"
)

// 对目标IP进行存活检测
var ping = &cli.Command{
	Name:  "ping",
	Usage: "target is timo to live",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "target",
			Aliases: []string{"t"},
			Usage:   "target domain",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "targets",
			Aliases: []string{"ts"},
			Usage:   "targets domain, use ',' split domain, explame: xxxx,xxxx,xxxx",
			Value:   "",
		},
	},
	Before: func(c *cli.Context) error {
		var err = errors.New("need target")
		target = c.String("target")
		if target != "" {
			return nil
		}
		target = c.String("targets")
		if target == "" {
			return err
		}
		targets = strings.Split(target, ",")
		return nil
	},
	Action: func(c *cli.Context) error {
		// 获取网关 Mac 地址
		mac := pkg.GetGatewayMac()

		var (
			dns       *protocol.DnsProtocol
			icmp      *protocol.IcmpProtocol
			d         *device.Device
			r         *runner.Runner
			err       error
			arpCtx    context.Context
			arpCancel context.CancelFunc
			freePort  int
			hosts     = make(map[string]bool)
			ipCh      = make(chan []byte)
			srcIp     []byte
			ts        [][]byte
		)
		if target != "" {
			ts = append(ts, []byte(target))
		}

		for _, s := range targets {
			ts = append(ts, []byte(s))
		}

		d = pkg.DefaultDevice()
		freePort, err = pkg.GetFreePort()
		if err != nil {
			ylog.WithField("command", "ping").Errorf(err.Error())
			return nil
		}
		srcIp = d.Ipv4.To4()
		arpCtx, arpCancel = context.WithTimeout(context.Background(), 15*time.Second)
		dns = protocol.NewDnsProtocol(arpCtx, arpCancel, freePort, protocol.WithDnsResolvers())

		icmpCtx, icmpCancel := context.WithTimeout(context.Background(), 15*time.Second)
		icmp = protocol.NewIcmpProtocol(icmpCtx, icmpCancel, protocol.WithBufferNumber(15))

		dnsPacketCh := dns.BuildSendPacket(srcIp, d.Mac, mac, ts)
		icmpPacketCh := icmp.BuildSendPacket(srcIp, d.Mac, mac, ipCh)

		r = pkg.DefaultRunner()
		if r.Err != nil {
			ylog.WithField("command", "ping").Errorf(err.Error())
			return nil
		}

		r.AppendParseHandle(func(packet gopacket.Packet) bool {
			return dns.Parse(packet)
		})
		r.AppendParseHandle(func(packet gopacket.Packet) bool {
			return icmp.Parse(packet)
		})

		go r.RunSender()
		go r.RunReceive()

		f := func(r *runner.Runner, packetCh <-chan []byte) {
			for packet := range packetCh {
				r.PushPacket(packet)
			}
		}
		icmpSend := func(ip []byte, ipCh chan []byte) {
			ipCh <- ip
		}

		go f(r, dnsPacketCh)
		go f(r, icmpPacketCh)

		for {
			select {
			case data := <-dns.DnsTable:
				ylog.WithFields(map[string]string{
					"command": "ping",
					"Target":  data.Host,
					"DnsType": data.DnsType.String(),
					"Value":   data.Value,
				}).Info("Finish")
				if data.DnsType == layers.DNSTypeA || data.DnsType == layers.DNSTypeAAAA {
					if _, ok := hosts[data.Value]; !ok {
						timer.RunTime(context.Background(), 5, 1*time.Second, func() {
							icmpSend(net.ParseIP(data.Value).To4(), ipCh)
						})
						hosts[data.Value] = true
						icmp.Add(5)
					}

				}
			case data := <-icmp.IcmpTable:
				ylog.WithFields(map[string]string{
					"command": "ping",
					"ICMP":    data.Target,
					"TTL":     string(data.Ttl),
					"Time":    data.Elapsed.String(),
				}).Info("ICMP to DNS")
			case <-icmpCtx.Done():
				ylog.WithField("command", "ping").Infof("timeout")
				r.DoneCh()
				r.Close()
				return nil
			}
		}
	},
}
