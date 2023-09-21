package device

import (
	"errors"
	"github.com/google/gopacket/pcap"
	"net"
	"strings"
)

type Device struct {
	Ipv4   net.IP
	IpMask net.IPMask
	Mac    net.HardwareAddr

	NetName  string
	NetFlags net.Flags

	PcapName        string
	PcapDescription string
	PcapFlags       uint32
	Err             error
}

func (s *Device) Collect() (err error) {
	s.netInfo()
	s.pcapInfo()
	return s.Err
}

// 搜集网络信息
func (s *Device) netInfo() {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	defer conn.Close()
	if err != nil {
		s.Err = err
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	s.Ipv4 = localAddr.IP.To4()
	s.IpMask = localAddr.IP.DefaultMask()

	devices, _ := net.Interfaces()
	for _, device := range devices {

		if !isUp(device) {
			continue
		}
		if isLoopback(device) {
			continue
		}
		// exclude VMware Network
		if strings.HasPrefix(strings.ToLower(device.Name), "vmware") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(device.Name), "vethernet") {
			continue
		}

		s.Mac = device.HardwareAddr
		s.NetFlags = device.Flags
		s.NetName = device.Name
		return

	}
	err = errors.New("can't find net device")
	s.Err = err
}

// 搜集网卡设备信息
func (s *Device) pcapInfo() {
	if s.Err != nil {
		return
	}
	var (
		devices []pcap.Interface
		err     error
	)

	devices, err = pcap.FindAllDevs()
	if err != nil {
		s.Err = err
		return
	}

	for _, device := range devices {
		for _, address := range device.Addresses {
			if address.IP.Equal(s.Ipv4) {
				s.PcapFlags = device.Flags
				s.PcapName = device.Name
				s.PcapDescription = device.Description
				return
			}
		}
	}
	err = errors.New("can not find pcap devices")
	s.Err = err
}

func isUp(d net.Interface) bool {
	return d.Flags&net.FlagUp != 0
}

func isLoopback(d net.Interface) bool {
	return d.Flags&net.FlagLoopback != 0
}
