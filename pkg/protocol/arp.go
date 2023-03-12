package protocol

import (
	"bytes"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	manuf "github.com/timest/gomanuf"
	"github.com/xtclalala/ylog"
	"net"
)

func NewArpProtocol() *ArpProtocol {

	return &ArpProtocol{
		ArpMap:   make(map[string]bool),
		ArpTable: make(chan ArpInfo),
	}
}

type ArpProtocol struct {
	ArpMap   map[string]bool
	ArpTable chan ArpInfo
	srcIp    []byte
	srcMac   []byte
}

func (s *ArpProtocol) BuildSendPacket(srcIp, srcMac []byte, dstIps []net.IP) <-chan []byte {
	var sendCh = make(chan []byte)
	s.srcIp = srcIp
	s.srcMac = srcMac

	go func(sendCh chan []byte) {
		var (
			opt    gopacket.SerializeOptions
			buffer gopacket.SerializeBuffer
			err    error
			eth    = &layers.Ethernet{}
			a      = &layers.ARP{}
		)
		defer close(sendCh)

		eth.SrcMAC = srcMac
		eth.DstMAC = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		eth.EthernetType = layers.EthernetTypeARP

		// 构建 arp 报
		a.AddrType = layers.LinkTypeEthernet // 硬件类型 以太网
		a.Protocol = layers.EthernetTypeIPv4 // 协议类型 ip地址
		a.HwAddressSize = 6                  // 固定硬件地址长度 6
		a.ProtAddressSize = 4                // 固定协议地址长度 4
		a.Operation = layers.ARPRequest      // 1:arp请求报文 2:arp应答报文
		a.SourceHwAddress = srcMac
		a.SourceProtAddress = srcIp
		a.DstHwAddress = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		buffer = gopacket.NewSerializeBuffer()
		for _, dstIp := range dstIps {

			// 构建 arp 报
			a.DstProtAddress = dstIp.To4()

			// 形成 bytes
			err = gopacket.SerializeLayers(buffer, opt, eth, a)
			if err != nil {
				ylog.WithField("command", "arp").Errorf("build packet buffer is failed")
				continue
			}

			sendCh <- buffer.Bytes()
			err = buffer.Clear()
			if err != nil {
				ylog.WithField("command", "arp").Errorf("clear packet buffer is failed")
				continue
			}
			s.ArpMap[dstIp.String()] = true
			ylog.WithField("command", "arp").Debugf("%s packet buffer is ok", dstIp.String())
		}
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
	if !bytes.Equal(arp.DstHwAddress, s.srcMac) || !bytes.Equal(arp.DstProtAddress, s.srcIp) {
		return false
	}

	srcIp := net.IP(arp.SourceProtAddress).String()
	srcMac := net.HardwareAddr(arp.SourceHwAddress)
	m := manuf.Search(srcMac.String())
	s.ArpTable <- NewArpInfo(srcIp, srcMac.String(), m)

	ylog.WithField("command", "arp").Debugf(srcIp)
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
