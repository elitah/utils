package netlib

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

var (
	ENoValidAddress = errors.New("no valid address")
)

func SplitHostPort(v interface{}) (string, int, error) {
	//
	if nil != v {
		//
		if s, ok := v.(string); ok {
			//
			host, port, err := net.SplitHostPort(s)
			//
			if "" != host {
				//
				if n, err := strconv.ParseInt(port, 10, 32); err == nil {
					//
					return host, int(n), nil
				}
				//
				return host, 0, nil
			}
			//
			return host, 0, err
		} else if addr, ok := v.(*net.IPAddr); ok {
			return addr.IP.String(), 0, nil
		} else if addr, ok := v.(*net.TCPAddr); ok {
			return addr.IP.String(), addr.Port, nil
		} else if addr, ok := v.(*net.UDPAddr); ok {
			return addr.IP.String(), addr.Port, nil
		} else if addr, ok := v.(*net.UnixAddr); ok {
			return addr.Name, 0, nil
		}
	}
	//
	return "", 0, ENoValidAddress
}

func GetIPMaskRoute() (string, string, string) {
	//
	if "linux" == strings.ToLower(runtime.GOOS) {
		//
		if f, err := os.Open("/proc/net/route"); nil == err {
			//
			defer f.Close()
			//
			if r := bufio.NewReader(f); nil != r {
				//
				splitFn := func(c rune) bool {
					//
					if unicode.IsLetter(c) {
						return false
					}
					//
					if unicode.IsNumber(c) {
						return false
					}
					//
					switch c {
					case '-', '_':
						//
						return false
					}
					//
					return true
				}
				//
				for {
					//
					if line, _, err := r.ReadLine(); nil == err {
						//
						if tags := strings.FieldsFunc(string(line), splitFn); 4 <= len(tags) {
							//
							if v, err := strconv.ParseInt(tags[3], 16, 32); nil == err {
								//
								if 0x3 == v {
									//
									if iface, err := net.InterfaceByName(tags[0]); nil == err {
										//
										if list, _ := iface.Addrs(); 0 < len(list) {
											//
											var ipaddr net.IP
											var ipmask net.IPMask
											//
											for _, item := range list {
												//
												ipaddr = nil
												//
												switch result := item.(type) {
												case *net.IPNet:
													//
													ipaddr = result.IP
													ipmask = result.Mask
												case *net.IPAddr:
													//
													ipaddr = result.IP
												}
												//
												if nil != ipaddr {
													//
													ipaddr = ipaddr.To4()
												}
												//
												if nil != ipaddr && ipaddr.IsGlobalUnicast() {
													//
													break
												}
											}
											//
											if nil != ipaddr {
												//
												var gateway string
												//
												if v, err := strconv.ParseInt(tags[2], 16, 64); nil == err {
													//
													gateway = fmt.Sprintf(
														"%d.%d.%d.%d",
														(v>>0)&0xFF,
														(v>>8)&0xFF,
														(v>>16)&0xFF,
														(v>>24)&0xFF,
													)
												}
												//
												if nil == ipmask {
													//
													ipmask = ipaddr.DefaultMask()
												}
												//
												return ipaddr.String(), net.IP(ipmask).String(), gateway
											}
										}
									}
								}
							}
						} else {
							//
							break
						}
					} else {
						//
						break
					}
				}
			}
		}
	}
	//
	return "", "", ""
}
