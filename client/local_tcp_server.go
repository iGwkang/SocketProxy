// +build !windows

package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"errors"
	"net"
	"os"
	"syscall"
	"time"
)

type LocalTCPServer struct {
	listenAddr  string
	serverAddrs []string
	encType     uint8
	password    string
	timeout     time.Duration
}

func NewLocalTCPServer(localAddr string, servers []string, encType uint8, passwd string, timeout time.Duration) *LocalTCPServer {
	return &LocalTCPServer{
		listenAddr:  localAddr,
		serverAddrs: servers,
		encType:     encType,
		password:    passwd,
		timeout:     timeout,
	}
}

func (c *LocalTCPServer) getServerConn(remoteAddr string) (net.Conn, error) {
	for i := 0; i < len(c.serverAddrs); i++ {
		conn, err := DialServer(c.serverAddrs[i], remoteAddr, c.password, c.encType)
		if err == nil {
			return conn, nil
		}
	}
	return nil, errors.New("server connection failed")
}

func (c *LocalTCPServer) Run() {
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

func (c *LocalTCPServer) TcpClientHandle(conn net.Conn) {
	defer conn.Close()

	if c.timeout != 0 {
		conn.(*net.TCPConn).SetKeepAlivePeriod(c.timeout)
	}

	// 获取目标地址
	addr, err := getSockDestAddr(conn)
	if err != nil {
		Logger.Warn(err)
		return
	}
	// 判断是否在白名单
	ip, port := common.AddrToString(addr)

	if !IPisProxy(ip) {
		serverConn, err := net.DialTimeout("tcp", ip+":"+port, c.timeout)
		if err != nil {
			Logger.Warn(err)
			return
		}
		defer serverConn.Close()

		if c.timeout != 0 {
			serverConn.(*net.TCPConn).SetKeepAlivePeriod(c.timeout)
		}
		Logger.Debugf("Start relay %s <--> %s", conn.RemoteAddr(), serverConn.RemoteAddr())
		common.Relay(serverConn, conn)
	} else {
		// 与服务器建立连接
		serverConn, err := c.getServerConn(ip + ":" + port)
		if err != nil {
			Logger.Warn(err)
			return
		}
		defer serverConn.Close()

		Logger.Debugf("start relay %s <--> %s <--> %s", conn.RemoteAddr(), serverConn.RemoteAddr(), ip+":"+port)
		common.Relay(serverConn, conn)
	}
}

// 获取目标地址
func getSockDestAddr(conn net.Conn) (addr []byte, err error) {
	const SO_ORIGINAL_DST = 80

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