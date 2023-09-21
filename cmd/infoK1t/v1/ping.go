package main

import (
	"errors"
	"github.com/google/gopacket/layers"
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/internal/config"
	"github.com/xtclalala/infoK1t/internal/output"
	"github.com/xtclalala/infoK1t/internal/timer"
	"github.com/xtclalala/infoK1t/pkg"
	"go.uber.org/zap"
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
			Usage:   "targets domain, use ',' split, explame:  baidu.com or baidu.com,bing.com,bilibili.com",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "dns server, use ',' split, explame: 22.22.22.22 or 11.11.11.11,33.33.33.33",
			Value:   "",
		},
	},
	Action: func(c *cli.Context) error {
		var (
			target  string
			targets []string

			server  string
			servers = []string{}
		)
		var err = errors.New("need target")
		target = c.String("target")
		if target == "" {
			return err
		}
		targets = strings.Split(target, ",")

		if c.String("server") != "" {
			servers = strings.Split(server, ",")
		}

		icmpSend := func(ip []byte, ipCh chan []byte) {
			ipCh <- ip
		}
		var (
			ts    [][]byte
			hosts = make(map[string]bool)
		)
		for _, t := range targets {
			ts = append(ts, []byte(t))
		}
		r := pkg.DefaultRunner()
		dns, icmp, err := pkg.Ping(ts, append(servers, config.GetOptions().Dns.Servers...))
		if err != nil {
			output.GetLogger().Error(err.Error())
			return nil
		}

		for {
			select {
			case data := <-dns.DnsTable:
				output.GetLogger().Info("DNS",
					zap.String("Target", data.Host),
					zap.String("DnsType", data.DnsType.String()),
					zap.String("Value", data.Value),
					zap.String("Server", data.SrcIp.String()),
				)
				if data.DnsType == layers.DNSTypeA || data.DnsType == layers.DNSTypeAAAA {
					// 去重
					if _, ok := hosts[data.Value]; !ok {
						timer.RunTime(r.Ctx, 5, 1*time.Second, func() {
							icmpSend(net.ParseIP(data.Value).To4(), icmp.DstIpCh)
						})
						hosts[data.Value] = true
						icmp.Add(5)
					}

				}
			case data := <-icmp.IcmpTable:
				output.GetLogger().Info("ICMP",
					zap.String("From", data.Target),
					zap.Uint8("TTL", data.Ttl),
					zap.String("Time", data.Elapsed.String()),
				)
			case <-time.After(10 * time.Second):
				output.GetLogger().Info("timeout")
				r.Close()
				return nil
			case <-icmp.Ctx.Done():
				output.GetLogger().Info("is done")
				r.Close()
				return nil
			}
		}
	},
}
