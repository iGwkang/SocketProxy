package main

import (
	"SocketProxy/common"
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
)

type serverConnect struct {
	net.Conn
	encType    uint8
	password   string
	remoteAddr string
}

func DialServer(addr, remoteAddr, passwd string, encType uint8) (c net.Conn, err error) {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	serverConn := &serverConnect{
		Conn:       conn,
		encType:    encType,
		password:   passwd,
		remoteAddr: remoteAddr,
	}

	if err = serverConn.handshake(); err != nil {
		serverConn.Close()
		return nil, err
	}
	return serverConn, nil
}

func (c *serverConnect) handshake() (err error) {
	_, err = c.Write([]byte{c.encType})
	if err != nil {
		return
	}

	switch c.encType {
	case 0: // 异或
		xorByte := common.GetNonZeroNumber()
		_, err = c.Write([]byte{xorByte})
		c.Conn = common.NewXorCipher(xorByte, c.Conn)
	case 1: // tls
		tlsConn := tls.Client(c.Conn, TLSConfig)
		err = tlsConn.Handshake()
		c.Conn = tlsConn
	default:
		err = errors.New("Encryption type not supported")
	}
	if err != nil {
		return
	}

	// 给服务器校验密码
	_, err = c.Write([]byte(c.password))

	sha := sha1.Sum([]byte(c.password))
	buf := make([]byte, len(sha))
	_, err = io.ReadFull(c, buf)
	if err != nil {
		return
	}
	// 校验服务器返回的密码
	if !bytes.Equal(sha[:], buf) {
		err = errors.New("Server Password Error")
		return
	}

	ip, port, err := net.SplitHostPort(c.remoteAddr)
	if err != nil {
		err = errors.New("remote addr Parse Error")
	}

	destAddr := make([]byte, 6)

	ipVal := common.Inet_addr(ip)
	portVal, _ := strconv.Atoi(port)

	binary.BigEndian.PutUint32(destAddr, ipVal)
	binary.BigEndian.PutUint16(destAddr[4:6], uint16(portVal))

	_, err = c.Write(destAddr)
	return
}
