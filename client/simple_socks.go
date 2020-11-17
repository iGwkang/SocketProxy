// +build windows

package main

import (
	"errors"
	"net"

	"github.com/miekg/dns"
)

func GetSockDestAddr(conn net.Conn) (addr []byte, err error) {
	msg := make([]byte, 512)
	_, err = conn.Read(msg)
	if err != nil {
		return
	}
	switch msg[0] {
	case 0x4:
		return GetSocks4DestAddr(msg, conn)
	case 0x5:
		return GetSocks5DestAddr(conn)
	default:
		err = errors.New("protocol not support")
	}
	return
}

// socks4
func GetSocks4DestAddr(msg []byte, conn net.Conn) (addr []byte, err error) {
	// TODO socks4
	err = errors.New("socks4 protocol not support")
	return nil, err
}

// socks5
func GetSocks5DestAddr(conn net.Conn) (addr []byte, err error) {
	// 客户端回应：Socks服务端不需要验证方式
	_, err = conn.Write([]byte{0x05, 0x00})
	if err != nil {
		return
	}

	var b [512]byte
	n, err := conn.Read(b[:])
	if err != nil {
		return
	}

	if b[0] != 0x5 || b[1] != 0x01 { //CONNECT
		err = errors.New("only support CONNECT")
		return
	}

	addr = make([]byte, 6)
	switch b[3] {
	case 0x01: //IP V4
		copy(addr, b[4:8])
	case 0x03: //域名
		domain := string(b[5 : n-2]) //b[4]表示域名的长度
		// dns解析
		ip := ParseDomainIPv4(domain)
		if ip == nil {
			err = errors.New("Parse Failed " + domain)
			return
		}
		copy(addr, ip)
	case 0x04: //IP V6
		err = errors.New("Not support ipv6")
		return
	}
	// 端口
	copy(addr[4:6], b[n-2:n])

	//响应客户端连接成功
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	return
}

func ParseDomainIPv4(domain string) []byte {
	if ip := net.ParseIP(domain); ip != nil {
		return ip.To4()
	}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	in, err := dns.Exchange(&m, "127.0.0.1:"+ClientConfig.ListenDNSPort)
	if err != nil {
		return nil
	}
	for _, ans := range in.Answer {
		if a, ok := ans.(*dns.A); ok {
			return a.A.To4()
		}
	}

	return nil
}
