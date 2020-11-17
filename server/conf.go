package main

import (
	. "SocketProxy/logger"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"
)

var configPath = flag.String("config", "config.json", "Config File Path.")

// 服务端配置
var ServerConfig = struct {
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

	err = json.Unmarshal(data, &ServerConfig)
	if err != nil {
		Logger.Warn(err)
	}
	ServerConfig.Timeout *= time.Second
	Logger.Infof("Server Config: %+v", ServerConfig)
}

func InitTLSConfig() {
	cert, err := generateCert()
	if err != nil {
		Logger.Fatal(err)
	}
	TLSConfig = &tls.Config{
		ServerName:   "SocketProxy",
		Certificates: []tls.Certificate{cert},
		VerifyConnection: func(cs tls.ConnectionState) error {
			if cs.ServerName != TLSConfig.ServerName {
				return fmt.Errorf("client server name is %s", cs.ServerName)
			}
			return nil
		},
		MinVersion: tls.VersionTLS12,
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

func generateCert() (cert tls.Certificate, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}

	template := x509.Certificate{SerialNumber: serialNumber}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	cert, err = tls.X509KeyPair(certPEM, keyPEM)
	return
}
