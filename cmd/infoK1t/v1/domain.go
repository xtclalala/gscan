package main

import (
	"bufio"
	"github.com/urfave/cli/v2"
	"github.com/xtclalala/infoK1t/internal/config"
	"github.com/xtclalala/infoK1t/internal/output"
	"github.com/xtclalala/infoK1t/internal/util"
	"github.com/xtclalala/infoK1t/pkg"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

var subDomain = &cli.Command{
	Name:  "subdomain",
	Usage: "find subdomain web site by dns protocol",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "target",
			Aliases:  []string{"t"},
			Usage:    "targets domain, use ',' split, explame:  baidu.com or baidu.com,bing.com,bilibili.com",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "server",
			Aliases:  []string{"s"},
			Usage:    "dns server, use ',' split, explame: 22.22.22.22 or 11.11.11.11,33.33.33.33",
			Required: false,
			Value:    "",
		},
		//&cli.IntFlag{
		//	Name:  "level",
		//	Usage: "域名层级深度。(&:建议不超过4)",
		//	Required: false,
		//	Value: 3,
		//},
	},
	Action: func(c *cli.Context) error {
		// todo 读字典，构建需要查找的域名
		r := pkg.DefaultRunner()
		var servers = []string{}
		targets := strings.Split(c.String("target"), ",")
		var target = make([][]byte, 0, 10000)
		if c.String("servers") != "" {
			servers = strings.Split(c.String("servers"), ",")
		}
		file, _ := os.Open("./doc/subdomain.txt")
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			data, err := util.Sanitize(scan.Text())
			if err != nil {
				output.GetLogger().Error(err.Error())
				continue
			}
			for _, t := range targets {
				t, err = util.Sanitize(t)
				if err != nil {
					output.GetLogger().Error(err.Error())
					continue
				}
				data = data + "." + t
				target = append(target, []byte(data))
			}
		}
		dns, err := pkg.Subdomain(target, append(servers, config.GetOptions().Dns.Servers...))
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
			case <-dns.Ctx.Done():
				output.GetLogger().Info("is done")
				r.Close()
				return nil
			case <-time.After(20 * time.Second):
				output.GetLogger().Info("timeout")
				r.Close()
				return nil
			}
		}
		return nil

	},
}
