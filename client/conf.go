package main

import (
	. "SocketProxy/logger"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"strings"
	"time"
)

var configPath = flag.String("config", "config.json", "Config File Path.")

// 客户端配置
var conf = struct {
	ListenTcpAddrs   string `json:"listen_tcp_addr"`    // 本地tcp代理
	ListenHttpAddr   string `json:"listen_http_addr"`   // 本地http代理
	ListenSocks5Addr string `json:"listen_socks5_addr"` // 本地socks5代理
	ListenDNSAddr    string `json:"listen_dns_addr"`    // 本地dns监听地址
	ListenDNSPort    string
	//ListenUDPAddr  string   `json:"listen_udp_addr"` // TODO: 本地udp监听地址
	TcpServerAddrs []string      `json:"tcp_servers"`   // tcp服务器地址
	DNSServer      string        `json:"dns_server"`    // vps dns服务器地址 (12.12.12.12:25353)
	Timeout        time.Duration `json:"timeout"`       // 超时时间
	Password       string        `json:"password"`      // 密码
	Encryption     uint8         `json:"encryption"`    // 加密方式 (0: 异或, 非0:tls)
	DomainFile     string        `json:"domain_file"`   // 域名文件
	IPWhiteFile    string        `json:"ip_white_file"` // ip白名单文件
}{
	ListenTcpAddrs:   "0.0.0.0:1080",
	ListenHttpAddr:   "0.0.0.0:1081",
	ListenSocks5Addr: "0.0.0.0:1082",
	ListenDNSAddr:    "0.0.0.0:25353",
	Timeout:          0,
	Encryption:       0,
	Password:         "SocketProxy",
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
	conf.ListenDNSPort = conf.ListenDNSAddr[strings.IndexByte(conf.ListenDNSAddr, ':')+1:]
	Logger.Infof("Client Config: %+v", conf)
	passwd := sha1.Sum([]byte(conf.Password))
	conf.Password = string(passwd[:])
}

func InitTLSConfig() {
	TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
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
