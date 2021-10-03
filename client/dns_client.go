package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"errors"
	"github.com/patrickmn/go-cache"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type DNSClient struct {
	listenAddr *net.UDPAddr
	listener   *net.UDPConn
	bytePool   sync.Pool
	dnsMsgPool sync.Pool
	dnsServer  string
	timeout    time.Duration
}

var lookupCache = cache.New(1*time.Hour, 10*time.Minute)

func NewDNSClient(addr, dnsServer string, timeout time.Duration) *DNSClient {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		Logger.Error(err)
		return nil
	}

	return &DNSClient{
		listenAddr: laddr,
		bytePool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 512)
			},
		},
		dnsMsgPool: sync.Pool{
			New: func() interface{} {
				return &dns.Msg{}
			},
		},
		dnsServer: dnsServer,
		timeout:   timeout,
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
		sendData = make([]byte, len(buf)+1)

		Logger.Debug(domain, " is proxy")
		xorByte := common.GetNonZeroNumber()
		for i := 0; i < len(buf); i++ {
			buf[i] ^= xorByte
		}
		sendData[0] = xorByte
		copy(sendData[1:], buf)

		sendData, err = common.RequestDNSParse(sendData, s.dnsServer, s.timeout)
		if err != nil {
			Logger.Error(err)
			return
		}

		for i := 0; i < len(sendData); i++ {
			sendData[i] ^= xorByte
		}

	} else {
		Logger.Debug(domain, " not proxy")
		sendData, err = common.RequestDNSParse(buf, conf.ChinaDNSServer+":53", conf.Timeout)
		if err != nil {
			Logger.Error(err)
			return
		}

		/*
			err = dnsMsg.Unpack(data)
			if err != nil {
				Logger.Error(err)
				return
			}

			for i := 0; i < len(dnsMsg.Answer); {
				if _, ok := dnsMsg.Answer[i].(*dns.AAAA); ok {
					dnsMsg.Answer = append(dnsMsg.Answer[:i], dnsMsg.Answer[i+1:]...)
				} else {
					i++
				}
			}
			sendData, err = dnsMsg.PackBuffer(data)
			if err != nil {
				Logger.Error(err)
				return
			}
		*/
	}

	_, err = s.listener.WriteToUDP(sendData, cliAddr)
	if err != nil {
		Logger.Error(err)
		return
	}
}

func ParseDomain(domain string) (string, error) {

	if ip, ok := lookupCache.Get(domain); ok {
		return ip.(string), nil
	}

	if ip := net.ParseIP(domain).To4(); ip != nil {
		return ip.String(), nil
	}

	if DomainIsProxy(domain) {
		m := dns.Msg{}
		m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

		in, err := dns.Exchange(&m, "127.0.0.1:"+conf.ListenDNSPort)
		if err != nil {
			return "", err
		}
		for _, ans := range in.Answer {
			if a, ok := ans.(*dns.A); ok {
				ip := a.A.To4().String()
				lookupCache.Set(domain, ip, 0)
				return ip, nil
			}
		}
		return "", errors.New("parse domain " + domain + " failure")
	} else {
		host, err := net.LookupHost(domain)
		if err != nil {
			return "", err
		}
		return host[0], nil
	}
}
