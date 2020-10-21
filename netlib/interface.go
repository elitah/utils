package netlib

import (
	"net"
)

func InterfaceCheckIPContains(ip string) bool {
	//
	if _ip := net.ParseIP(ip); nil != _ip {
		if list, err := InterfaceAddrs(); nil == err {
			for _, item := range list {
				switch result := item.(type) {
				case *net.IPNet:
					if result.Contains(_ip) {
						return true
					}
				case *net.IPAddr:
					if result.IP.Equal(_ip) {
						return true
					}
				case *net.TCPAddr:
					if result.IP.Equal(_ip) {
						return true
					}
				case *net.UDPAddr:
					if result.IP.Equal(_ip) {
						return true
					}
				case *net.UnixAddr:
				}
			}
		}
	}
	//
	return false
}

func InterfaceAddrs() ([]net.Addr, error) {
	return InterfaceAddrsFilter(false, false)
}

func InterfaceAddrsFilter(mustUp, noLoopback bool) ([]net.Addr, error) {
	//
	var results []net.Addr
	//
	if list, err := net.Interfaces(); nil == err {
		//
		for _, item := range list {
			//
			if mustUp && 0 == (net.FlagUp&item.Flags) {

				continue
			}
			//
			if noLoopback && 0 != (net.FlagLoopback&item.Flags) {
				continue
			}
			//
			if addrs, err := item.Addrs(); nil == err {
				results = append(results, addrs...)
			}
		}
	} else {
		//
		return nil, err
	}
	//
	return results, nil
}
