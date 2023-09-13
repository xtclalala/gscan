package protocol

import "github.com/google/gopacket"

type IProtocol[T any] interface {
	BuildSendPacket() <-chan []byte
	Parse(packet gopacket.Packet) bool
}
