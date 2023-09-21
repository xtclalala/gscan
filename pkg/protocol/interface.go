package protocol

import (
	"context"
	"github.com/google/gopacket"
)

type IProtocol[T any] interface {
	BuildSendPacket(ctx context.Context, handle func(T)) <-chan []byte
	Parse(packet gopacket.Packet) bool
}
