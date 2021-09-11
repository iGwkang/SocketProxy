package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"net"
	"sync"
	"time"
)

type TcpServer struct {
	listenAddrs []string
}

func NewTcpServer() *TcpServer {
	return &TcpServer{
		listenAddrs: conf.ListenTcpAddrs,
	}
}

func (s *TcpServer) Run() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.listenAddrs))
	for i := 0; i < len(s.listenAddrs); i++ {
		go func(addr string) {
			defer wg.Done()
			TcpServerListenAddr(addr)
		}(s.listenAddrs[i])

	}
	wg.Wait()
}

func TcpServerListenAddr(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		Logger.Error(err)
		return err
	}
	defer listener.Close()

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			Logger.Error(err)
			break
		}

		go TcpServerHandle(conn)
	}
	return err
}

func TcpServerHandle(conn net.Conn) {
	if conf.Timeout != 0 {
		_ = conn.(*net.TCPConn).SetKeepAlivePeriod(conf.Timeout)
		conn.SetDeadline(time.Now().Add(conf.Timeout))
	}

	newConn, ip, port, cipherType, err := Handshake(conn)
	if err != nil {
		Logger.Warn("Remote Addr: ", conn.RemoteAddr(), " Handshake error: ", err)
		conn.Close()
		return
	}
	defer newConn.Close()

	if conf.Timeout != 0 {
		conn.SetDeadline(time.Time{})
	}

	// 访问目标地址
	dstConn, err := net.DialTimeout("tcp", ip+":"+port, conf.Timeout)
	if err != nil {
		Logger.Error(err)
		return
	}
	defer dstConn.Close()
	if conf.Timeout != 0 {
		_ = dstConn.(*net.TCPConn).SetKeepAlivePeriod(conf.Timeout)
	}
	Logger.Debugf("Use cipherType: %#v, start relay %s <--> %s", cipherType, newConn.RemoteAddr(), dstConn.RemoteAddr())
	common.Relay(dstConn, newConn)
}
