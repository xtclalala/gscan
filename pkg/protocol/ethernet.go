package protocol

import "github.com/google/gopacket/layers"

type EthernetProtocol struct {
	layers.Ethernet
}

func (s *EthernetProtocol) BuildEthernet() *layers.Ethernet {
	return &layers.Ethernet{
		SrcMAC:       s.SrcMAC,
		DstMAC:       s.DstMAC,
		EthernetType: s.EthernetType,
	}
}
