package device

import (
	"errors"
	"github.com/google/gopacket/pcap"
	"github.com/xtclalala/ylog"
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
		ylog.WithField("command", "device").Debugf(err.Error())
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ylog.WithField("command", "device").Debugf(localAddr.String())
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

		//var addrs []net.Addr
		_, err = device.Addrs()
		if err != nil {
			ylog.WithField("command", "device").Debugf(err.Error())
			s.Err = err
			return
		}

		s.Mac = device.HardwareAddr
		s.NetFlags = device.Flags
		s.NetName = device.Name
		return

	}
	err = errors.New("can not find net devices")
	s.Err = err
	ylog.WithField("command", "device").Debugf(err.Error())
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
		ylog.WithField("command", "device").Debugf(err.Error())
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
	ylog.WithField("command", "device").Debugf(err.Error())
}

func isUp(d net.Interface) bool {
	return d.Flags&net.FlagUp != 0
}

func isLoopback(d net.Interface) bool {
	return d.Flags&net.FlagLoopback != 0
}

func getIpv4AndMask(addrs []net.Addr) (ip net.IP, mask *net.IPNet, err error) {
	var temp net.Addr
	if len(addrs) == 2 {
		temp = addrs[1]
	} else {
		temp = addrs[0]
	}
	ip, mask, err = net.ParseCIDR(temp.String())
	return
}
