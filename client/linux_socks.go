// +build linux

package main

import (
	"errors"
	"net"
	"os"
	"syscall"
)

const SO_ORIGINAL_DST = 80

func GetSockDestAddr(conn net.Conn) (addr []byte, err error) {
	var file *os.File
	switch conn.(type) {
	case *net.TCPConn:
		// 获取客户端需要访问的地址
		file, err = conn.(*net.TCPConn).File()
		if err != nil {
			return
		}
	case *net.UDPConn:
		file, err = conn.(*net.UDPConn).File()
		if err != nil {
			return
		}
	default:
		err = errors.New("Conn not support.")
		return
	}
	defer file.Close()

	// 返回sockaddr_in
	addr6, err := syscall.GetsockoptIPv6Mreq(int(file.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		return
	}

	addr = make([]byte, 6)
	copy(addr, addr6.Multiaddr[4:8])
	copy(addr[4:6], addr6.Multiaddr[2:4])
	return
}
