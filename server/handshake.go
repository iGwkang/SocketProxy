package main

import (
	"SocketProxy/common"
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"io"
	"net"
)

// 握手
func Handshake(conn net.Conn) (newConn net.Conn, ip, port string, cipherType uint16, err error) {
	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if n <= 0 || err != nil {
		return nil, "", "", 0, err
	}

	// 加密类型
	switch buf[0] {
	case 0:
		newConn, err = xorHandshake(conn)
	case 1:
		newConn, cipherType, err = tlsHandshake(conn)
	default:
		err = errors.New("encryption type not supported")
	}
	if err != nil {
		return
	}
	// 密码验证
	err = verifyClientPassword(newConn)
	if err != nil {
		return
	}

	// 给客户端回哈希之后的密码
	sha := sha1.Sum([]byte(ServerConfig.Password))
	newConn.Write(sha[:])

	// 目的地址
	ip, port, err = getDestAddr(newConn)
	if err != nil {
		return
	}

	// 判断是否在内网
	if common.IsIntranetAddress(ip) {
		err = errors.New(ip + " is IntranetAddress")
		return
	}
	return
}

func xorHandshake(conn net.Conn) (newConn net.Conn, err error) {
	xorByte := [1]byte{}
	n, err := conn.Read(xorByte[:])
	if n <= 0 || err != nil {
		return nil, errors.New("xorHandshake error")
	}

	if xorByte[0] == 0 {
		return nil, errors.New("xorHandshake error, xorByte == 0")
	}
	newConn, _ = common.NewXorCipher(xorByte[0], conn)
	return
}

func tlsHandshake(conn net.Conn) (newConn net.Conn, cipherType uint16, err error) {
	tlsConn := tls.Server(conn, TLSConfig)
	newConn = tlsConn
	err = tlsConn.Handshake()
	cipherType = tlsConn.ConnectionState().CipherSuite
	return
}

func verifyClientPassword(conn net.Conn) (err error) {
	passwd := make([]byte, len(ServerConfig.Password))
	// 密码校验
	_, err = io.ReadFull(conn, passwd)
	if err != nil {
		return err
	}
	if string(passwd) != ServerConfig.Password {
		return errors.New("password verification failed")
	}
	return
}

func getDestAddr(conn net.Conn) (ip, port string, err error) {
	addrByte := [6]byte{}
	n, err := io.ReadFull(conn, addrByte[:])
	if n <= 0 || err != nil {
		return "", "", err
	}

	ip, port = common.AddrToString(addrByte[:])
	return
}
