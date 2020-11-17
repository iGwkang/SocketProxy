package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"crypto/tls"
	"errors"
	"net"
)

type TcpClient struct {
	listenAddr  string
	serverAddrs []string
}

func NewTcpClient() *TcpClient {
	return &TcpClient{
		listenAddr:  ClientConfig.ListenTcpAddrs,
		serverAddrs: ClientConfig.TcpServerAddrs,
	}
}

func (c *TcpClient) getServerConn() (net.Conn, error) {
	for i := 0; i < len(c.serverAddrs); i++ {
		conn, err := net.DialTimeout("tcp", c.serverAddrs[i], ClientConfig.Timeout)
		if err == nil {
			return conn, err
		}
	}
	return nil, errors.New("server connection failed")
}

func (c *TcpClient) Run() {
	if len(c.serverAddrs) == 0 {
		Logger.Error("TcpServer number = 0")
		return
	}

	listener, err := net.Listen("tcp", c.listenAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
	defer listener.Close()

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			Logger.Error(err)
			break
		}
		go c.TcpClientHandle(conn)

	}
	return
}

func (c *TcpClient) TcpClientHandle(conn net.Conn) {
	defer conn.Close()

	if ClientConfig.Timeout != 0 {
		conn.(*net.TCPConn).SetKeepAlivePeriod(ClientConfig.Timeout)
	}

	// 获取目标地址
	addr, err := GetSockDestAddr(conn)
	if err != nil {
		Logger.Warn(err)
		return
	}
	// 判断是否在白名单
	ip, port := common.AddrToString(addr)

	if !IPisProxy(ip) {
		serverConn, err := net.DialTimeout("tcp", ip+":"+port, ClientConfig.Timeout)
		if err != nil {
			Logger.Warn(err)
			return
		}
		defer serverConn.Close()

		if ClientConfig.Timeout != 0 {
			serverConn.(*net.TCPConn).SetKeepAlivePeriod(ClientConfig.Timeout)
		}
		Logger.Debugf("start relay %s <--> %s", conn.RemoteAddr(), serverConn.RemoteAddr())
		common.Relay(serverConn, conn)
	} else {
		// 与服务器建立连接
		serverConn, err := c.getServerConn()
		if err != nil {
			Logger.Warn(err)
			return
		}
		defer func() {
			serverConn.Close()
		}()

		serverConn, cipherType, err := c.Handshake(serverConn, addr)
		if err != nil {
			Logger.Warn(err)
			return
		}
		Logger.Debugf("use cipherType: %#v, start relay %s <--> %s <--> %s", cipherType, conn.RemoteAddr(), serverConn.RemoteAddr(), ip+":"+port)
		common.Relay(serverConn, conn)
	}
}

// 与服务端握手
func (c *TcpClient) Handshake(serConn net.Conn, destAddr []byte) (newConn net.Conn, cipherType uint16, err error) {
	_, err = serConn.Write([]byte{ClientConfig.Encryption})
	switch ClientConfig.Encryption {
	case 0: // 异或
		xorByte := common.GetNonZeroNumber()
		_, err = serConn.Write([]byte{xorByte})
		newConn, _ = common.NewXorCipher(xorByte, serConn)
		cipherType = 0x00
	default: // tls
		tlsConn := tls.Client(serConn, TLSConfig)
		newConn = tlsConn
		err = tlsConn.Handshake()
		cipherType = tlsConn.ConnectionState().CipherSuite
	}
	_, err = newConn.Write([]byte(ClientConfig.Password))
	_, err = newConn.Write(destAddr[:])
	return
}
