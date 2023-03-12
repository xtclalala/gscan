package protocol

import (
	"context"
	"github.com/xtclalala/ylog"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/xtclalala/infoK1t/internal/id"
	"net"
	"sync"
	"time"
)

type IcmpOptionFunc func(protocol *IcmpProtocol)

func WithBufferNumber(number int) IcmpOptionFunc {
	if number < 10 {
		number = 10
	}
	return func(protocol *IcmpProtocol) {
		protocol.IcmpTable = make(chan IcmpInfo, number)
	}
}

func NewIcmpProtocol(c context.Context, cancel context.CancelFunc, optionFunc ...IcmpOptionFunc) *IcmpProtocol {
	var p = &IcmpProtocol{
		Ctx:            c,
		Cancel:         cancel,
		Cache:          map[uint16]IcmpInfo{},
		PacketCount:    0,
		CountIsRunning: false,
		IcmpType:       layers.CreateICMPv4TypeCode(8, 0),
	}

	for _, f := range optionFunc {
		f(p)
	}
	return p
}

type IcmpProtocol struct {
	Cache          map[uint16]IcmpInfo
	IcmpTable      chan IcmpInfo
	PacketCount    uint
	CountIsRunning bool
	mx             sync.Mutex
	Ctx            context.Context
	Cancel         context.CancelFunc
	IcmpType       layers.ICMPv4TypeCode
}

func (s *IcmpProtocol) BuildSendPacket(srcIp, srcMac []byte, dstMac net.HardwareAddr, dstIpCh <-chan []byte) <-chan []byte {
	var sendCh = make(chan []byte)

	go func(sendCh chan []byte) {
		var (
			opt    gopacket.SerializeOptions
			buffer gopacket.SerializeBuffer
			err    error
			eth    = &layers.Ethernet{}
			ip     = &layers.IPv4{}
			icmp   = &layers.ICMPv4{}
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
		ip.TTL = 128
		ip.Protocol = layers.IPProtocolICMPv4
		ip.SrcIP = srcIp

		icmp.TypeCode = s.IcmpType
		buffer = gopacket.NewSerializeBuffer()

		opt.FixLengths = true
		opt.ComputeChecksums = true

		for dstIp := range dstIpCh {
			i := NewIcmpInfo()
			icmp.Id = i.Id
			icmp.Seq = i.Seq
			// 构建 ip 报
			ip.DstIP = dstIp
			s.mx.Lock()
			s.Cache[i.Id] = i
			s.mx.Unlock()
			// 形成 bytes
			err = gopacket.SerializeLayers(buffer, opt, eth, ip, icmp)
			if err != nil {
				ylog.WithField("command", "icmp").Errorf("build packet buffer is failed: %s", err.Error())
				continue
			}
			sendCh <- buffer.Bytes()
			err = buffer.Clear()
			if err != nil {
				ylog.WithField("command", "icmp").Errorf("clear packet buffer is failed")
				continue
			}
			ylog.WithField("command", "icmp").Debugf(string(dstIp))

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
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	ip := ipLayer.(*layers.IPv4)
	s.mx.Lock()
	i := s.Cache[icmp.Id]
	s.mx.Unlock()
	i.WithEnd(ip.TTL, ip.SrcIP.String())
	s.IcmpTable <- i
	s.Done()
	if s.CountIsRunning && s.PacketCount == 0 {
		s.Cancel()
	}
	return true
}

func (s *IcmpProtocol) Add(c uint) {
	s.CountIsRunning = true
	s.mx.Lock()
	defer s.mx.Unlock()
	s.PacketCount += c
}

func (s *IcmpProtocol) Done() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.PacketCount -= 1
}

func NewIcmpInfo() IcmpInfo {
	return IcmpInfo{
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
