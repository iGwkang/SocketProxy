package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"net"
	"sync"
)

type TcpServer struct {
	listenAddrs []string
}

func NewTcpServer() *TcpServer {
	return &TcpServer{
		listenAddrs: ServerConfig.ListenTcpAddrs,
	}
}

func (s *TcpServer) Run() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.listenAddrs))
	for i := 0; i < len(s.listenAddrs); i++ {
		go func(addr string) {
			_ = TcpServerListenAddr(addr)
			wg.Done()
		}(s.listenAddrs[i])

	}
	wg.Wait()
	return
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
	if ServerConfig.Timeout != 0 {
		_ = conn.(*net.TCPConn).SetKeepAlivePeriod(ServerConfig.Timeout)
	}

	newConn, ip, port, err := Handshake(conn)
	if err != nil {
		conn.Close()
		return
	}
	defer newConn.Close()
	Logger.Debug("request:", ip, ":", port)
	// 访问目标地址
	dstConn, err := net.DialTimeout("tcp", ip+":"+port, ServerConfig.Timeout)
	if err != nil {
		Logger.Error(err)
		return
	}
	defer dstConn.Close()
	if ServerConfig.Timeout != 0 {
		_ = dstConn.(*net.TCPConn).SetKeepAlivePeriod(ServerConfig.Timeout)
	}

	common.Relay(dstConn, newConn)
}
