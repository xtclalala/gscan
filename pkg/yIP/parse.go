package yIP

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

var (
	ErrNetMaskTooBig = errors.New("NetMask is too big")
	ErrIsNotIp       = errors.New("is not ip")
)

func ParseCIDR(ipStr string) (ips []int, err error) {
	var (
		ip    int
		mask  int
		ipNet *net.IPNet
	)
	_, ipNet, err = net.ParseCIDR(ipStr)
	ip, err = Ip2int(ipNet.IP.To4().String())
	if err != nil {
		return
	}
	mask, _ = ipNet.Mask.Size()
	return Parse(ip, mask)
}

func ParseString(ipStr, maskStr string) (ips []int, err error) {
	var (
		ip   int
		mask int
	)
	ip, err = Ip2int(ipStr)
	if err != nil {
		return
	}
	mask, err = strconv.Atoi(maskStr)
	if err != nil {
		return
	}
	return Parse(ip, mask)
}

// Parse 解析 IP 返回该 IP 所在网段的所有IP 不包括网络地址
func Parse(ip int, mask int) (ips []int, err error) {
	if mask >= 32 {
		err = ErrNetMaskTooBig
		return
	}
	ips = make([]int, 0, 1<<(32-mask)-2)
	var networkNumber int
	var hostNumber int

	networkNumber = ip >> (32 - mask) << (32 - mask)
	hostNumber = 1 << (32 - mask)
	for i := 1; i < hostNumber; i++ {
		ips = append(ips, networkNumber+i)
	}
	return
}

func Ip2int(ip string) (iip int, err error) {
	if !IsIp(ip) {
		err = ErrIsNotIp
		return
	}
	for index, node := range strings.Split(ip, ".") {
		temp, e := strconv.Atoi(node)
		if e != nil {
			err = e
			return
		}
		iip += temp << ((3 - index) * 8)
	}
	return
}

func Int2Ip(iip int) (ip string) {
	var temps = make([]string, 4)
	var temp int
	for i := 1; i <= 4; i++ {
		temp = iip & 0xFF
		temps[4-i] = strconv.Itoa(temp)
		iip >>= 8

	}
	return strings.Join(temps, ".")
}

func IsIp(ip string) bool {
	if net.ParseIP(ip) != nil {
		return true
	}
	return false

}
