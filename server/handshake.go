package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"crypto/tls"
	"errors"
	"io"
	"net"
)

// 握手
func Handshake(conn net.Conn) (newConn net.Conn, ip, port string, err error) {
	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if n <= 0 || err != nil {
		return nil, "", "", errors.New("Handshake Error")
	}

	// 加密类型
	switch buf[0] {
	case 0:
		newConn, err = xorHandshake(conn)
	default:
		newConn, err = tlsHandshake(conn)
	}
	if err != nil {
		Logger.Warn("Handshake Error:", err)
		return
	}
	// 密码验证
	err = verifyClientPassword(newConn)
	if err != nil {
		newConn.Close()
		Logger.Warn(err)
		return
	}

	// 目的地址
	ip, port, err = getDestAddr(newConn)
	if err != nil {
		newConn.Close()
		Logger.Warn(err)
		return
	}

	// 判断是否在内网
	if common.IsIntranetAddress(ip) {
		newConn.Close()
		Logger.Warn(ip, "is IntranetAddress!")
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

func tlsHandshake(conn net.Conn) (newConn net.Conn, err error) {
	newConn = tls.Server(conn, TLSConfig)
	return
}

func verifyClientPassword(conn net.Conn) (err error) {
	passwd := make([]byte, len(ServerConfig.Password))
	// 密码校验
	_, err = io.ReadFull(conn, passwd)
	if err != nil {
		return errors.New("password verification failed")
	}
	if string(passwd) != ServerConfig.Password {
		return errors.New("password verification failed, password:" + string(passwd))
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
