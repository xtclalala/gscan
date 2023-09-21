package pkg

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/xtclalala/infoK1t/pkg/device"
	"github.com/xtclalala/infoK1t/pkg/protocol"
	"github.com/xtclalala/infoK1t/pkg/runner"
	"github.com/xtclalala/infoK1t/pkg/yIP"
	"net"
	"strings"
	"sync"
)

var (
	r    *runner.Runner
	once sync.Once
)

func DefaultRunner() *runner.Runner {
	once.Do(func() {
		d = DefaultDevice()
		option := runner.NewOptions(d.PcapName, 1024, false, 10, "!src "+d.Ipv4.String())
		r = runner.NewRunner(*option)
		r.Open()
	})

	return r
}

func Identify() (*device.Device, error) {
	d = DefaultDevice()
	return d, d.Err
}

func Probe(target string) (arpInfoCh chan protocol.ArpInfo, err error) {
	var (
		arp *protocol.ArpProtocol
	)
	d = DefaultDevice()

	var (
		srcIp     = d.Ipv4.To4()
		srcIpInt  int
		maskSize  int
		targetIps []int
		temps     = make([]net.IP, 0, 255)
	)
	if target == "" {
		// 不传参 对当前网段进行设备扫描
		srcIpInt, err = yIP.Ip2int(srcIp.String())
		if err != nil {
			return
		}
		maskSize, _ = d.IpMask.Size()
		targetIps, err = yIP.Parse(srcIpInt, maskSize)
		if err != nil {
			return
		}
		for _, ip := range targetIps {
			temps = append(temps, net.ParseIP(yIP.Int2Ip(ip)))
		}
		arp = protocol.NewArpProtocol(func(protocol *protocol.ArpProtocol) {
			protocol.DstMAC, protocol.SrcMAC, protocol.EthernetType = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, d.Mac, layers.EthernetTypeARP
		})

	} else {
		targets := strings.Split(target, ",")
		for _, t := range targets {
			targetIps, err = yIP.ParseCIDR(t)
			if err != nil {
				return
			}
			for _, ip := range targetIps {
				temps = append(temps, net.ParseIP(yIP.Int2Ip(ip)))
			}
		}

		arp = protocol.NewArpProtocol(func(protocol *protocol.ArpProtocol) {
			protocol.DstMAC, protocol.SrcMAC, protocol.EthernetType = GetGatewayMac(), d.Mac, layers.EthernetTypeARP
		})

	}
	arpInfoCh = arp.ArpTable

	r = DefaultRunner()
	if r.Err != nil {
		return
	}

	r.AppendParseHandle(func(packet gopacket.Packet) bool {
		return arp.Parse(packet)
	})

	go r.RunSender()
	go r.RunReceive()
	packetCh := arp.BuildSendPacket(r.Ctx, func(ap *protocol.ArpProtocol) {
		ap.SrcIp, ap.DstIps = srcIp, temps
	})

	for packet := range packetCh {
		r.PushPacket(packet)
	}
	return

}

func Ping(targets [][]byte, servers []string) (*protocol.DnsProtocol, *protocol.IcmpProtocol, error) {

	// 获取网关 Mac 地址
	mac = GetGatewayMac()

	var (
		dns      *protocol.DnsProtocol
		icmp     *protocol.IcmpProtocol
		err      error
		freePort int
		srcIp    []byte
		ipCh     = make(chan []byte)
	)

	d = DefaultDevice()
	r = DefaultRunner()

	if err != nil {
		return nil, nil, err
	}
	srcIp = d.Ipv4.To4()
	dns = protocol.NewDnsProtocol(protocol.WithDnsResolvers(servers...))

	icmp = protocol.NewIcmpProtocol(protocol.WithBufferNumber(15))
	freePort, err = GetFreePort()
	if err != nil {
		return nil, nil, err
	}
	dnsPacketCh := dns.BuildSendPacket(r.Ctx, func(dp *protocol.DnsProtocol) {
		dp.HostCh = targets
		// ethernet
		dp.SrcMAC = d.Mac
		dp.DstMAC = mac
		dp.EthernetType = layers.EthernetTypeIPv4

		dp.SrcIP = srcIp
		// udp
		dp.SrcPort = layers.UDPPort(freePort)
		dp.DstPort = layers.UDPPort(53)
	})
	icmpPacketCh := icmp.BuildSendPacket(r.Ctx, func(ip *protocol.IcmpProtocol) {
		// ethernet
		ip.SrcMAC = d.Mac
		ip.DstMAC = mac
		ip.EthernetType = layers.EthernetTypeIPv4
		// ip4
		ip.SrcIP = srcIp
		ip.DstIpCh = ipCh
	})

	r = DefaultRunner()
	if r.Err != nil {
		return dns, icmp, r.Err
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
	go f(r, dnsPacketCh)
	go f(r, icmpPacketCh)
	return dns, icmp, err
}

func Subdomain(targets [][]byte, servers []string) (*protocol.DnsProtocol, error) {
	mac = GetGatewayMac()
	d = DefaultDevice()
	freePort, err := GetFreePort()
	if err != nil {
		return nil, err
	}
	dns := protocol.NewDnsProtocol(protocol.WithDnsResolvers(servers...))
	dnsPacketCh := dns.BuildSendPacket(r.Ctx, func(dp *protocol.DnsProtocol) {
		dp.HostCh = targets
		// ethernet
		dp.SrcMAC = d.Mac
		dp.DstMAC = mac
		dp.EthernetType = layers.EthernetTypeIPv4

		dp.SrcIP = d.Ipv4.To4()
		// udp
		dp.SrcPort = layers.UDPPort(freePort)
		dp.DstPort = layers.UDPPort(53)
	})

	r = DefaultRunner()
	if r.Err != nil {
		return nil, r.Err
	}
	r.AppendParseHandle(func(packet gopacket.Packet) bool {
		return dns.Parse(packet)
	})

	for packet := range dnsPacketCh {
		r.PushPacket(packet)
	}
	return dns, nil
}
