package protocol

import (
	"bytes"
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	manuf "github.com/timest/gomanuf"
	"net"
)

type ArpOptionFunc func(protocol *ArpProtocol)

func NewArpProtocol(optionFunc ...ArpOptionFunc) *ArpProtocol {
	arp := &ArpProtocol{
		ArpMap:   make(map[string]bool),
		ArpTable: make(chan ArpInfo),
	}
	for _, f := range optionFunc {
		f(arp)
	}
	return arp
}

type ArpProtocol struct {
	EthernetProtocol
	ArpMap   map[string]bool
	ArpTable chan ArpInfo
	SrcIp    []byte
	DstIps   []net.IP
	Err      error
}

func (s *ArpProtocol) BuildSendPacket(ctx context.Context, f func(protocol *ArpProtocol)) <-chan []byte {
	f(s)
	var sendCh = make(chan []byte)
	srcIp := s.SrcIp
	dstIps := s.DstIps
	_, cancel := context.WithCancel(ctx)
	go func(sendCh chan []byte) {
		var (
			opt    gopacket.SerializeOptions
			buffer gopacket.SerializeBuffer
			eth    *layers.Ethernet
			a      = &layers.ARP{}
		)
		defer close(sendCh)

		eth = s.BuildEthernet()

		// 构建 arp 报
		a.AddrType = layers.LinkTypeEthernet // 硬件类型 以太网
		a.Protocol = layers.EthernetTypeIPv4 // 协议类型 ip地址
		a.HwAddressSize = 6                  // 固定硬件地址长度 6
		a.ProtAddressSize = 4                // 固定协议地址长度 4
		a.Operation = layers.ARPRequest      // 1:arp请求报文 2:arp应答报文
		a.SourceHwAddress = s.SrcMAC
		a.SourceProtAddress = srcIp
		a.DstHwAddress = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		buffer = gopacket.NewSerializeBuffer()

		for _, dstIp := range dstIps {

			// 构建 arp 报
			a.DstProtAddress = dstIp.To4()

			// 形成 bytes
			s.Err = gopacket.SerializeLayers(buffer, opt, eth, a)
			if s.Err != nil {
				return
			}

			sendCh <- buffer.Bytes()
			s.Err = buffer.Clear()
			if s.Err != nil {
				return
			}
			s.ArpMap[dstIp.String()] = true
		}
		cancel()
	}(sendCh)

	return sendCh
}

func (s *ArpProtocol) Parse(packet gopacket.Packet) bool {
	layer := packet.Layer(layers.LayerTypeARP)
	if layer == nil {
		return false
	}

	arp := layer.(*layers.ARP)
	if arp == nil {
		return false
	}

	// 过滤其他设备的请求报文
	if arp.Operation == layers.ARPRequest {
		return false
	}

	// 过滤掉目标ip和mac不是本机的报文
	if !bytes.Equal(arp.DstHwAddress, s.SrcMAC) || !bytes.Equal(arp.DstProtAddress, s.SrcIp) {
		return false
	}

	srcIp := net.IP(arp.SourceProtAddress).String()
	srcMac := net.HardwareAddr(arp.SourceHwAddress)
	m := manuf.Search(srcMac.String())
	s.ArpTable <- NewArpInfo(srcIp, srcMac.String(), m)

	return true
}

func NewArpInfo(ip, mac, m string) ArpInfo {
	return ArpInfo{
		ip, mac, m,
	}
}

type ArpInfo struct {
	Ip  string
	Mac string
	M   string
}
