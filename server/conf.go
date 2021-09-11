package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"time"
)

var configPath = flag.String("config", "config.json", "Config File Path.")

// 服务端配置
var conf = struct {
	ListenTcpAddrs []string      `json:"listen_tcp_addrs"` // 监听地址 (一个服务器可能有多个ip地址)
	ListenDNSAddr  string        `json:"listen_dns_addr"`  // dns监听地址
	DNSServer      string        `json:"dns_server"`       // dns服务器地址 (8.8.8.8:53)
	Timeout        time.Duration `json:"timeout"`          // 超时时间
	Password       string        `json:"password"`         // 密码
}{
	ListenTcpAddrs: []string{"0.0.0.0:23456"},
	ListenDNSAddr:  "0.0.0.0:25353",
	DNSServer:      "8.8.8.8:53",
	Timeout:        5,
	Password:       "SocketProxy",
}

var TLSConfig *tls.Config

func LoadConfig() {
	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		Logger.Warn(err)
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		Logger.Warn(err)
	}
	conf.Timeout *= time.Second
	if conf.Timeout == 0 {
		conf.Timeout = 30 * time.Second
	}

	Logger.Infof("Server Config: %+v", conf)

	passwd := sha1.Sum([]byte(conf.Password))
	conf.Password = string(passwd[:])
}

func InitTLSConfig() {
	certPEM, keyPEM, err := common.GenerateCert()
	if err != nil {
		Logger.Fatal(err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		Logger.Fatal(err)
	}

	TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}
}
