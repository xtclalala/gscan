package protocol

import "github.com/google/gopacket/layers"

type UdpProtocol struct {
	layers.UDP
}

func (s *UdpProtocol) BuildUdp(ip *layers.IPv4) *layers.UDP {
	udp := &layers.UDP{
		SrcPort: s.SrcPort,
		DstPort: layers.UDPPort(53),
	}
	_ = udp.SetNetworkLayerForChecksum(ip)
	return udp
}
