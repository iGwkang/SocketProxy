package main

import (
	"SocketProxy/common"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"net"
)

type serverConnect struct {
	net.Conn
	encType  uint8
	password string
	dstip    string
	dstport  uint16
}

func DialServer(proxyServerAddr, dstip string, dstport uint16, passwd string, encType uint8) (c net.Conn, err error) {
	conn, err := net.Dial("tcp", proxyServerAddr)
	if err != nil {
		return nil, err
	}
	serverConn := &serverConnect{
		Conn:     conn,
		encType:  encType,
		password: passwd,
		dstip:    dstip,
		dstport:  dstport,
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
		return errors.New("Encryption type not supported")
	}
	if err != nil {
		return
	}
	buf := make([]byte, len(c.password) + 6)
	copy(buf, common.StringToBytes(c.password))
	binary.BigEndian.PutUint32(buf[len(c.password):], common.Inet_addr(c.dstip))
	binary.BigEndian.PutUint16(buf[len(c.password) + 4:], c.dstport)
	_, err = c.Write(buf)
	return
}
