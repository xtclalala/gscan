package protocol

import "github.com/google/gopacket/layers"

type Ip4Protocol struct {
	layers.IPv4
}

func (s *Ip4Protocol) BuildIp4(protocol layers.IPProtocol) *layers.IPv4 {
	return &layers.IPv4{
		Version:    4,
		IHL:        5,
		TOS:        0,
		Length:     0,
		Checksum:   0,
		Id:         0,
		Flags:      layers.IPv4DontFragment,
		FragOffset: 0,
		TTL:        128,
		Protocol:   protocol,
		SrcIP:      s.SrcIP,
	}
}
