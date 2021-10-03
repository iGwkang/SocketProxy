package common

import (
	"encoding/binary"
	"net"
	"net/http"
	"strconv"
	"unsafe"
)

type IPRange struct {
	Begin uint32
	End   uint32
}

// 内网地址范围
var intranetAddr = [...]IPRange{
	{Begin: Inet_addr("127.0.0.0"), End: Inet_addr("127.255.255.255")},
	{Begin: Inet_addr("10.0.0.0"), End: Inet_addr("10.255.255.255")},
	{Begin: Inet_addr("169.254.0.0"), End: Inet_addr("169.254.255.255")},
	{Begin: Inet_addr("172.16.0.0"), End: Inet_addr("172.31.255.255")},
	{Begin: Inet_addr("192.168.0.0"), End: Inet_addr("192.168.255.255")},
}

// ip地址转为uint32
func Inet_addr(ipv4 string) uint32 {
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
	addr := Inet_addr(ipv4)

	for i := 0; i < len(intranetAddr); i++ {
		if addrInRange(addr, intranetAddr[i]) {
			return true
		}
	}

	return false
}

// 地址转为字符串
func AddrToString(addr []byte) (ip, port string) {
	if len(addr) != 6 {
		return
	}

	ip += strconv.FormatInt(int64(addr[0]), 10) + "."
	ip += strconv.FormatInt(int64(addr[1]), 10) + "."
	ip += strconv.FormatInt(int64(addr[2]), 10) + "."
	ip += strconv.FormatInt(int64(addr[3]), 10)
	port = strconv.FormatInt(int64(binary.BigEndian.Uint16(addr[4:6])), 10)
	return
}

// 获取外网ip
func GetExternalIP() net.IP {
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	buf := make([]byte, 32)

	n, err := resp.Body.Read(buf)
	if err != nil {
		return nil
	}
	ip := string(buf[:n])
	return net.ParseIP(ip).To4()
}
