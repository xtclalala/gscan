package protocol

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/xtclalala/infoK1t/internal/id"
	"sync"
	"time"
)

type IcmpOptionFunc func(protocol *IcmpProtocol)

func WithBufferNumber(number int) IcmpOptionFunc {
	if number < 10 {
		number = 10
	}
	return func(protocol *IcmpProtocol) {
		protocol.IcmpTable = make(chan *IcmpInfo, number)
	}
}

func NewIcmpProtocol(optionFunc ...IcmpOptionFunc) *IcmpProtocol {
	var p = &IcmpProtocol{
		Cache:       map[uint16]*IcmpInfo{},
		PacketCount: 0,
		IcmpType:    layers.CreateICMPv4TypeCode(8, 0),
	}

	for _, f := range optionFunc {
		f(p)
	}
	return p
}

type IcmpProtocol struct {
	EthernetProtocol
	Ip4Protocol
	Cache       map[uint16]*IcmpInfo
	IcmpTable   chan *IcmpInfo
	PacketCount uint
	mx          sync.Mutex
	Ctx         context.Context
	Cancel      context.CancelFunc
	IcmpType    layers.ICMPv4TypeCode
	DstIpCh     chan []byte
	Err         error
}

func (s *IcmpProtocol) BuildSendPacket(ctx context.Context, f func(protocol *IcmpProtocol)) <-chan []byte {
	f(s)
	var sendCh = make(chan []byte)
	s.Ctx, s.Cancel = context.WithCancel(ctx)
	go func(sendCh chan []byte) {
		var (
			opt    gopacket.SerializeOptions
			buffer gopacket.SerializeBuffer
			eth    *layers.Ethernet
			ip     *layers.IPv4
			icmp   = &layers.ICMPv4{}
		)
		defer close(sendCh)
		eth = s.BuildEthernet()
		ip = s.BuildIp4(layers.IPProtocolICMPv4)

		icmp.TypeCode = s.IcmpType
		buffer = gopacket.NewSerializeBuffer()

		opt.FixLengths = true
		opt.ComputeChecksums = true

		for dstIp := range s.DstIpCh {
			i := NewIcmpInfo()
			icmp.Id = i.Id
			icmp.Seq = i.Seq
			// 构建 ip 报
			ip.DstIP = dstIp
			s.mx.Lock()
			s.Cache[i.Id] = i
			s.mx.Unlock()
			// 形成 bytes
			s.Err = gopacket.SerializeLayers(buffer, opt, eth, ip, icmp)
			if s.Err != nil {
				return
			}
			sendCh <- buffer.Bytes()
			s.Err = buffer.Clear()
			if s.Err != nil {
				return
			}
		}
	}(sendCh)

	return sendCh
}

func (s *IcmpProtocol) Parse(packet gopacket.Packet) bool {
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer == nil {
		return false
	}

	icmp := icmpLayer.(*layers.ICMPv4)
	if icmp == nil {
		return false
	}
	if icmp.TypeCode == s.IcmpType {
		return false
	}

	s.mx.Lock()
	i, ok := s.Cache[icmp.Id]
	delete(s.Cache, icmp.Id)
	s.mx.Unlock()
	if !ok {
		return false
	}
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	ip := ipLayer.(*layers.IPv4)
	s.Done()

	i.WithEnd(ip.TTL, ip.SrcIP.String())
	s.IcmpTable <- i
	if s.PacketCount == 0 {
		s.Cancel()
	}
	return true
}

func (s *IcmpProtocol) Add(c uint) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.PacketCount += c
}

func (s *IcmpProtocol) Done() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.PacketCount -= 1
}

func NewIcmpInfo() *IcmpInfo {
	return &IcmpInfo{
		Id:        id.New("icmp"),
		Seq:       id.New("icmp"),
		StartTime: time.Now(),
	}
}

type IcmpInfo struct {
	Target    string
	StartTime time.Time
	Elapsed   time.Duration
	Id        uint16
	Seq       uint16
	Ttl       uint8
}

func (s *IcmpInfo) WithEnd(ttl uint8, target string) *IcmpInfo {
	if &(s.Elapsed) != nil {
		s.Elapsed = time.Now().Sub(s.StartTime)
		s.Ttl = ttl
		s.Target = target
	}

	return s
}
