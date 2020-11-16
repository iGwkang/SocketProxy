package common

import "unsafe"

type IPRange struct {
	Begin uint32
	End   uint32
}

// 内网地址范围
var intranetAddr = []IPRange{
	{Begin: Inet_addr_h("127.0.0.0"), End: Inet_addr_h("127.255.255.255")},
	{Begin: Inet_addr_h("10.0.0.0"), End: Inet_addr_h("10.255.255.255")},
	{Begin: Inet_addr_h("169.254.0.0"), End: Inet_addr_h("169.254.255.255")},
	{Begin: Inet_addr_h("172.16.0.0"), End: Inet_addr_h("172.31.255.255")},
	{Begin: Inet_addr_h("192.168.0.0"), End: Inet_addr_h("192.168.255.255")},
}

// ip地址转为uint32
func Inet_addr_h(ipv4 string) uint32 {
	var ret uint32
	var i int8 = 3
	var ptr *uint8

	for j := 0; j < len(ipv4) && i >= 0; j++ {
		if ipv4[j] == '.' {
			i--
		} else {
			ptr = (*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&ret)) + uintptr(i)))
			*ptr *= 10
			*ptr += ipv4[j] - '0'
		}
	}
	return ret
}

func addrInRange(addr uint32, r IPRange) bool {
	return (addr >= r.Begin && addr <= r.End)
}

// 判断是否是局域网ip
func IsIntranetAddress(ipv4 string) bool {
	addr := Inet_addr_h(ipv4)

	for i := 0; i < len(intranetAddr); i++ {
		if addrInRange(addr, intranetAddr[i]) {
			return true
		}
	}

	return false
}

