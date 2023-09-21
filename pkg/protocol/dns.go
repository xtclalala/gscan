package protocol

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	identify "github.com/xtclalala/infoK1t/internal/id"
	"github.com/zoumo/goset"
	"net"
)

type DnsOptionFunc func(protocol *DnsProtocol)

func WithDnsResolvers(ips ...string) DnsOptionFunc {
	return func(protocol *DnsProtocol) {
		protocol.DnsResolvers = goset.NewSafeSetFromStrings(ips).ToStrings()
	}
}

func NewDnsProtocol(optionFunc ...DnsOptionFunc) *DnsProtocol {
	var p = &DnsProtocol{
		DnsTable:     make(chan *DnsInfo),
		DnsResolvers: []string{},
		Cache:        make(map[string]*DnsInfo, 1000),
	}

	for _, f := range optionFunc {
		f(p)
	}
	return p
}

type DnsProtocol struct {
	EthernetProtocol
	Ip4Protocol
	UdpProtocol
	HostCh       [][]byte
	DnsResolvers []string
	DnsTable     chan *DnsInfo
	Cache        map[string]*DnsInfo
	Ctx          context.Context
	Cancel       context.CancelFunc
	Err          error
}

func (s *DnsProtocol) BuildSendPacket(ctx context.Context, f func(config *DnsProtocol)) <-chan []byte {
	f(s)
	s.Ctx, s.Cancel = context.WithCancel(ctx)
	var sendCh = make(chan []byte)

	go func(sendCh chan []byte) {
		var (
			opt    gopacket.SerializeOptions
			buffer gopacket.SerializeBuffer
			eth    *layers.Ethernet
			ip     *layers.IPv4
			udp    *layers.UDP
			dns    = &layers.DNS{}
		)
		defer close(sendCh)

		eth = s.BuildEthernet()
		ip = s.BuildIp4(layers.IPProtocolUDP)

		udp = s.BuildUdp(ip)

		dns.QDCount = 1
		dns.RD = true

		opt.FixLengths = true
		opt.ComputeChecksums = true

		buffer = gopacket.NewSerializeBuffer()
		for _, host := range s.HostCh {
			s.Cache[string(host)] = &DnsInfo{Host: string(host)}
			dns.Questions = []layers.DNSQuestion{
				{
					Name:  host,
					Type:  layers.DNSTypeA,
					Class: layers.DNSClassIN,
				},
			}
			for _, dnsIp := range s.DnsResolvers {
				// 构建 dns 报
				ip.DstIP = net.ParseIP(dnsIp).To4()
				id := identify.New("DNS")
				dns.ID = id
				// 形成 bytes
				s.Err = gopacket.SerializeLayers(buffer, opt, eth, ip, udp, dns)
				if s.Err != nil {
					return
				}

				sendCh <- buffer.Bytes()
				s.Err = buffer.Clear()
				if s.Err != nil {
					return
				}
			}
		}
	}(sendCh)

	return sendCh
}

func (s *DnsProtocol) Parse(packet gopacket.Packet) bool {
	dnsLayer := packet.Layer(layers.LayerTypeDNS)

	if dnsLayer == nil {
		return false
	}

	dns := dnsLayer.(*layers.DNS)
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
	ip4Layer := packet.Layer(layers.LayerTypeIPv4)

	if ip4Layer == nil {
		return true
	}

	ip4 := ip4Layer.(*layers.IPv4)
	if ip4 == nil {
		return true
	}
	value, ok := s.haveHost(dns.Questions[0].Name)
	if !ok {
		return true
	}

	identify.Del("DNS", dns.ID)
	for _, answer := range dns.Answers {
		if answer.Class != layers.DNSClassIN {
			continue
		}
		switch answer.Type {
		case layers.DNSTypeA, layers.DNSTypeAAAA:
			if answer.IP != nil {
				s.DnsTable <- WithValue(value, answer.Type, answer.IP.String(), ip4.SrcIP)
			}
		case layers.DNSTypeNS:
			if answer.NS != nil {
				s.DnsTable <- WithValue(value, answer.Type, string(answer.NS), ip4.SrcIP)
			}
		case layers.DNSTypeCNAME:
			if answer.CNAME != nil {
				s.DnsTable <- WithValue(value, answer.Type, string(answer.CNAME), ip4.SrcIP)
			}
		case layers.DNSTypePTR:
			if answer.PTR != nil {
				s.DnsTable <- WithValue(value, answer.Type, string(answer.PTR), ip4.SrcIP)
			}
		case layers.DNSTypeTXT:
			if answer.TXT != nil {
				s.DnsTable <- WithValue(value, answer.Type, string(answer.TXT), ip4.SrcIP)
			}
		}
	}

	if identify.Length("DNS") == 0 {
		s.Cancel()
	}
	return true
}

func (s *DnsProtocol) haveHost(host []byte) (*DnsInfo, bool) {
	i, ok := s.Cache[string(host)]
	return i, ok
}

type DnsInfo struct {
	Host    string
	DnsType layers.DNSType
	Value   string
	SrcIp   net.IP
}

func WithValue(dns *DnsInfo, dnsType layers.DNSType, value string, srcIp net.IP) *DnsInfo {
	dns.Value += "," + value
	dns.DnsType = dnsType
	dns.SrcIp = srcIp
	return dns
}
