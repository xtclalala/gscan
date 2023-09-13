package protocol

import "github.com/google/gopacket/layers"

type BaseProtocol struct {
	SrcMAC       []byte
	DstMAC       []byte
	EthernetType layers.EthernetType
}

func (s *BaseProtocol) Ethernet() *layers.Ethernet {
	return &layers.Ethernet{
		SrcMAC:       s.SrcMAC,
		DstMAC:       s.DstMAC,
		EthernetType: s.EthernetType,
	}
}
