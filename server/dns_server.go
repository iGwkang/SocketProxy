package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"net"
	"sync"
)

type DNSServer struct {
	listenAddr *net.UDPAddr
	listener   *net.UDPConn
	bytePool   sync.Pool
}

func NewDNSServer() *DNSServer {
	laddr, err := net.ResolveUDPAddr("udp", conf.ListenDNSAddr)
	if err != nil {
		Logger.Error(err)
		return nil
	}

	return &DNSServer{
		listenAddr: laddr,
		bytePool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024)
			},
		},
	}
}

// 服务端dns转发
func (s *DNSServer) Run() {
	var err error
	s.listener, err = net.ListenUDP("udp", s.listenAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
	defer s.listener.Close()

	for {
		buf := s.bytePool.Get().([]byte)

		n, cliAddr, err := s.listener.ReadFromUDP(buf) // 客户端连接
		if err != nil {
			Logger.Error(err)
			break
		}

		go func() {
			defer s.bytePool.Put(buf)
			s.handleDNS(buf[:n], cliAddr)
		}()
	}
}

func (s *DNSServer) handleDNS(buf []byte, cliAddr *net.UDPAddr) {
	if buf[0] == 0 {
		return
	}

	// 异或解密数据
	for i := 1; i < len(buf); i++ {
		buf[i] ^= buf[0]
	}

	data, err := common.RequestDNSParse(buf[1:], conf.DNSServer, conf.Timeout)
	if err != nil {
		Logger.Error(err)
		return
	}

	// 加密数据直接回给客户端
	for i := 0; i < len(data); i++ {
		data[i] ^= buf[0]
	}
	_, err = s.listener.WriteToUDP(data, cliAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
}
