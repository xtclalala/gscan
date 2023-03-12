package protocol

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/xtclalala/ylog"
	"net"
)

type DnsOptionFunc func(protocol *DnsProtocol)

func WithDnsResolvers(ips ...string) DnsOptionFunc {
	resolvers := []string{
		"223.5.5.5",
		"223.6.6.6",
		"119.29.29.29",
		"182.254.116.116",
		"114.114.114.115",
	}
	var rs = make([]string, 0, len(resolvers)+len(ips))
	rs = append(rs, resolvers...)
	rs = append(rs, ips...)
	return func(protocol *DnsProtocol) {
		protocol.DnsResolvers = rs
	}
}

func NewDnsProtocol(c context.Context, cancel context.CancelFunc, srcPort int, optionFunc ...DnsOptionFunc) *DnsProtocol {
	var p = &DnsProtocol{
		Ctx:          c,
		Cancel:       cancel,
		Cache:        map[string]DnsInfo{},
		DnsTable:     make(chan DnsInfo),
		SrcPort:      srcPort,
		DnsResolvers: []string{},
	}

	for _, f := range optionFunc {
		f(p)
	}
	return p
}

type DnsProtocol struct {
	DnsResolvers []string
	Cache        map[string]DnsInfo
	DnsTable     chan DnsInfo
	Ctx          context.Context
	Cancel       context.CancelFunc
	SrcPort      int
}

func (s *DnsProtocol) BuildSendPacket(srcIp, srcMac []byte, dstMac net.HardwareAddr, hosts [][]byte) <-chan []byte {
	var sendCh = make(chan []byte)

	go func(sendCh chan []byte) {
		var (
			opt    gopacket.SerializeOptions
			buffer gopacket.SerializeBuffer
			err    error
			eth    = &layers.Ethernet{}
			ip     = &layers.IPv4{}
			udp    = &layers.UDP{}
			dns    = &layers.DNS{}
		)
		defer close(sendCh)

		eth.SrcMAC = srcMac
		eth.DstMAC = dstMac
		eth.EthernetType = layers.EthernetTypeIPv4

		// 构建 ip 报
		ip.Version = 4 // ip 版本
		ip.IHL = 5
		ip.TOS = 0
		ip.Length = 0
		ip.Checksum = 0
		ip.Id = 0
		ip.Flags = layers.IPv4DontFragment
		ip.FragOffset = 0
		ip.TTL = 255
		ip.Protocol = layers.IPProtocolUDP
		ip.SrcIP = srcIp

		udp.SrcPort = layers.UDPPort(s.SrcPort)
		udp.DstPort = layers.UDPPort(53)

		dns.ID = 1
		dns.QDCount = 1
		dns.RD = true

		opt.FixLengths = true
		opt.ComputeChecksums = true

		_ = udp.SetNetworkLayerForChecksum(ip)

		buffer = gopacket.NewSerializeBuffer()
		for _, host := range hosts {
			dns.Questions = []layers.DNSQuestion{
				{
					Name:  host,
					Type:  layers.DNSTypeA,
					Class: layers.DNSClassIN,
				},
			}

			for _, dnsIp := range s.DnsResolvers {

				// 构建 arp 报
				ip.DstIP = net.ParseIP(dnsIp).To4()

				// 形成 bytes
				err = gopacket.SerializeLayers(buffer, opt, eth, ip, udp, dns)
				if err != nil {
					ylog.WithField("command", "dns").Errorf("build packet buffer is failed")
					continue
				}

				sendCh <- buffer.Bytes()
				err = buffer.Clear()
				if err != nil {
					ylog.WithField("command", "dns").Errorf("clear packet buffer is failed")
					continue
				}
				s.Cache[string(host)] = NewDnsInfo(string(host))
				ylog.WithField("command", "dns").Debugf("%s packet buffer is ok ", string(host))
			}
		}
	}(sendCh)

	return sendCh
}

func (s *DnsProtocol) Parse(packet gopacket.Packet) bool {
	layer := packet.Layer(layers.LayerTypeDNS)
	if layer == nil {
		return false
	}

	dns := layer.(*layers.DNS)
	if dns == nil {
		return false
	}

	// 过滤没有回答的请求报文
	if len(dns.Answers) == 0 {
		return true
	}

	if dns.ANCount <= 0 {
		return true
	}
	d := s.Cache[string(dns.Questions[0].Name)]
	for _, answer := range dns.Answers {
		if answer.Class != layers.DNSClassIN {
			continue
		}
		switch answer.Type {
		case layers.DNSTypeA, layers.DNSTypeAAAA:
			if answer.IP != nil {
				s.DnsTable <- d.WithValue(answer.Type, answer.IP.String())
			}
		case layers.DNSTypeNS:
			if answer.NS != nil {
				s.DnsTable <- d.WithValue(answer.Type, string(answer.NS))
			}
		case layers.DNSTypeCNAME:
			if answer.CNAME != nil {
				s.DnsTable <- d.WithValue(answer.Type, string(answer.CNAME))
			}
		case layers.DNSTypePTR:
			if answer.PTR != nil {
				s.DnsTable <- d.WithValue(answer.Type, string(answer.PTR))
			}
		case layers.DNSTypeTXT:
			if answer.TXT != nil {
				s.DnsTable <- d.WithValue(answer.Type, string(answer.TXT))
			}
		}
	}
	delete(s.Cache, d.Host)
	switch {
	case len(s.Cache) == 0:
		s.Cancel()
	}
	ylog.WithField("command", "dns").Debugf(d.Host)
	return true
}

func NewDnsInfo(host string) DnsInfo {
	return DnsInfo{
		Host: host,
	}
}

type DnsInfo struct {
	Host    string
	DnsType layers.DNSType
	Value   string
}

func (s *DnsInfo) WithValue(dnsType layers.DNSType, value string) DnsInfo {
	return DnsInfo{
		Host:    s.Host,
		Value:   value,
		DnsType: dnsType,
	}
}
