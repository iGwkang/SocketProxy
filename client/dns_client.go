package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"net"
	"sync"

	"github.com/miekg/dns"
)

type DNSClient struct {
	listenAddr *net.UDPAddr
	listener   *net.UDPConn
	bytePool   sync.Pool
	dnsMsgPool sync.Pool
}

func NewDNSClient() *DNSClient {
	laddr, err := net.ResolveUDPAddr("udp", ClientConfig.ListenDNSAddr)
	if err != nil {
		Logger.Error(err)
		return nil
	}

	return &DNSClient{
		listenAddr: laddr,
		bytePool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024)
			},
		},
		dnsMsgPool: sync.Pool{
			New: func() interface{} {
				return &dns.Msg{}
			},
		},
	}
}

func (c *DNSClient) Run() {
	var err error
	c.listener, err = net.ListenUDP("udp", c.listenAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
	defer c.listener.Close()

	for {
		buf := c.bytePool.Get().([]byte)

		n, cliAddr, err := c.listener.ReadFromUDP(buf) // 客户端连接
		if err != nil {
			Logger.Error(err)
			break
		}

		go func() {
			defer c.bytePool.Put(buf)
			c.handleDNS(buf[:n], cliAddr)
		}()
	}
}

func (s *DNSClient) handleDNS(buf []byte, cliAddr *net.UDPAddr) {
	dnsMsg := s.dnsMsgPool.Get().(*dns.Msg)
	defer s.dnsMsgPool.Put(dnsMsg)

	err := dnsMsg.Unpack(buf)
	if err != nil {
		Logger.Error(err)
		return
	}
	if len(dnsMsg.Question) == 0 {
		Logger.Error("dnsMsg.Question is NULL")
		return
	}
	domain := dnsMsg.Question[0].Name
	domain = domain[:len(domain)-1]

	var sendData []byte

	// 判断域名是否在黑名单
	if DomainIsProxy(domain) {
		xorByte := common.GetNonZeroNumber()
		for i := 0; i < len(buf); i++ {
			buf[i] ^= xorByte
		}
		sendData = []byte{xorByte}
		sendData = append(sendData, buf...)

		sendData, err = common.RequestDNSParse(sendData, ClientConfig.DNSServer, ClientConfig.Timeout)
		if err != nil {
			Logger.Error(err)
			return
		}

		for i := 0; i < len(sendData); i++ {
			sendData[i] ^= xorByte
		}

	} else {
		data, err := common.RequestDNSParse(buf, ClientConfig.CNDNSServer, ClientConfig.Timeout)
		if err != nil {
			Logger.Error(err)
			return
		}
		sendData = data
	}

	_, err = s.listener.WriteToUDP(sendData, cliAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
}
