package pkg

import (
	"github.com/google/gopacket"
	"github.com/xtclalala/infoK1t/pkg/device"
	"github.com/xtclalala/infoK1t/pkg/protocol"
	"github.com/xtclalala/infoK1t/pkg/runner"
	"github.com/xtclalala/infoK1t/pkg/yIP"
	"net"
	"sync"
)

var (
	d              *device.Device
	deviceInfoOnce sync.Once
)

func DefaultDevice() *device.Device {
	deviceInfoOnce.Do(func() {
		d = &device.Device{}
		d.Collect()
	})
	return d
}

var (
	mac         net.HardwareAddr
	GatewayOnce sync.Once
)

func GetGatewayMac() net.HardwareAddr {
	GatewayOnce.Do(func() {
		var (
			rr       *runner.Runner
			arp      *protocol.ArpProtocol
			srcIpInt int
			maskSize int
			ipi      int
		)
		rr = DefaultRunner()
		arp = protocol.NewArpProtocol()

		srcIpInt, _ = yIP.Ip2int(d.Ipv4.To4().String())
		maskSize, _ = d.IpMask.Size()
		ipi, _ = yIP.GatewayIp(srcIpInt, maskSize)

		packetCh := arp.BuildSendPacket(d.Ipv4.To4(), d.Mac, []net.IP{net.ParseIP(yIP.Int2Ip(ipi))})
		rr.AppendParseHandle(func(packet gopacket.Packet) bool {
			return arp.Parse(packet)
		})
		rr.Open()
		go rr.RunReceive()
		go rr.RunSender()
		packet := <-packetCh
		rr.PushPacket(packet)
		data := <-arp.ArpTable
		mac, _ = net.ParseMAC(data.Mac)

		rr.DoneCh()
	})
	return mac
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
