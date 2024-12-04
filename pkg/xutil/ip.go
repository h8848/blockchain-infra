package xutil

import (
	"net"
	"net/http"

	"github.com/thinkeridea/go-extend/exnet"
)

func ExternalIP() net.IP {
	iFaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iFace := range iFaces {
		if iFace.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iFace.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addresses, err := iFace.Addrs()
		if err != nil {
			return nil
		}
		for _, addr := range addresses {
			ip := GetIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip
		}
	}
	return nil
}

func GetIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}

// GetUserIp 获取客户端IP
func GetUserIp(r *http.Request) string {
	ip := exnet.ClientPublicIP(r)
	if ip == "" {
		ip = exnet.ClientIP(r)
	}
	return ip
}
