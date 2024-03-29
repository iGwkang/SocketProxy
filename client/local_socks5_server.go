package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"
)

type LocalSocks5Server struct {
	listenAddr  string
	serverAddrs []string
	encType     uint8
	password    string
	timeout     time.Duration
}

func NewLocalSocks5Server(localAddr string, servers []string, encType uint8, passwd string, timeout time.Duration) *LocalSocks5Server {
	return &LocalSocks5Server{
		listenAddr:  localAddr,
		serverAddrs: servers,
		encType:     encType,
		password:    passwd,
		timeout:     timeout,
	}
}

func (lss *LocalSocks5Server) getServerConn(ip string, port uint16) (net.Conn, error) {
	for i := 0; i < len(lss.serverAddrs); i++ {
		conn, err := DialServer(lss.serverAddrs[i], ip, port, lss.password, lss.encType)
		if err == nil {
			return conn, nil
		}
	}
	return nil, errors.New("server connection failed")
}

func (lss *LocalSocks5Server) Run() {
	if len(lss.serverAddrs) == 0 {
		Logger.Error("TcpServer number = 0")
		return
	}

	listener, err := net.Listen("tcp", lss.listenAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
	Logger.Info("start listen local socks5 server ", lss.listenAddr)

	defer listener.Close()

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			Logger.Error(err)
			break
		}
		go lss.TcpClientHandle(conn)
	}
}

func (lss *LocalSocks5Server) TcpClientHandle(conn net.Conn) {
	defer conn.Close()

	if lss.timeout != 0 {
		conn.(*net.TCPConn).SetKeepAlivePeriod(lss.timeout)
	}

	// 获取目标地址
	ip, port, err := getSock5Addr(conn)
	if err != nil {
		Logger.Warn(err)
		return
	}

	if !IPisProxy(ip) {
		serverConn, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(int(port)), lss.timeout)
		if err != nil {
			Logger.Warn(err)
			return
		}
		defer serverConn.Close()

		if lss.timeout != 0 {
			serverConn.(*net.TCPConn).SetKeepAlivePeriod(lss.timeout)
		}
		Logger.Debugf("Start relay %s <--> %s", conn.RemoteAddr(), serverConn.RemoteAddr())
		common.Relay(serverConn, conn)
	} else {
		// 与服务器建立连接
		serverConn, err := lss.getServerConn(ip, port)
		if err != nil {
			Logger.Warn(err)
			return
		}
		defer serverConn.Close()

		Logger.Debugf("start relay %v <--> %v <--> %v:%v", conn.RemoteAddr(), serverConn.RemoteAddr(), ip, port)
		common.Relay(serverConn, conn)
	}
}

func getSock5Addr(conn net.Conn) (ip string, port uint16, err error) {
	msg := make([]byte, 512)
	_, err = conn.Read(msg)
	if err != nil {
		return
	}
	switch msg[0] {
	case 0x5:
		return getSocks5DstAddr(conn)
	default:
		err = errors.New("Protocol not support")
	}
	return
}

// socks5
func getSocks5DstAddr(conn net.Conn) (ip string, port uint16, err error) {
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
		err = errors.New("Only support CONNECT")
		return
	}

	//addr = make([]byte, 6)
	switch b[3] {
	case 0x01: //IP V4
		ip = net.IP(b[4:8]).String()
	case 0x03: //域名
		domain := string(b[5 : n-2]) //b[4]表示域名的长度
		// dns解析
		ip, err = ParseDomain(domain)
		if err != nil {
			return
		}

	case 0x04: //IP V6
		err = errors.New("Not support ipv6")
		return
	}
	// 端口
	port = binary.BigEndian.Uint16(b[n-2 : n])

	//响应客户端连接成功
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	return
}
