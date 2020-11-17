package main

import (
	. "SocketProxy/logger"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
}

var TLSConfig *tls.Config

var certPem = []byte(`-----BEGIN CERTIFICATE-----
MIIDVzCCAj+gAwIBAgIJAIvw7Gzg91NcMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNV
BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQg
Q29tcGFueSBMdGQwHhcNMjAxMTE1MTQ0MTM0WhcNMzAxMTEzMTQ0MTM0WjBCMQsw
CQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZh
dWx0IENvbXBhbnkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
xPaFgXrchLFs5mZ3sZj4e4Ba/gOlOf1nNBAjNJ9hnXoKbIILT+5+GT349LZB57iZ
j9+FM2+Bs2Y57kJ/957oLPfzlC/PXsveiutoWV3zeY7+binqLMYEWgmP+wjunZ8t
Dj4k0EJr2PsHZlx73ST+dOUVVwvOQl7c2vgrHN8pSc+5z9Qk8RWGvBQm4Trk1Ciu
d7v6nc0Uh45WQwLdMJqRycsK+MBUHpe8TXkbCh4quscMw34pxgWAg/hfenLot2U9
qLW5zrhgCBP95h4hw0aulZXBB/mxADiyYElKLsBF7mgnC7ZIQmxmycbiaflWLwD/
Iei+9gKfVMI/8kzNKdf4LQIDAQABo1AwTjAdBgNVHQ4EFgQUZ3cOz4RFnYA9wYm2
tMYZSJrl+80wHwYDVR0jBBgwFoAUZ3cOz4RFnYA9wYm2tMYZSJrl+80wDAYDVR0T
BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAvB4dTIZ16d6sog0Sk7pjw2GuPHak
jHsDEF2Ml7jc5eD4liSUm/pX4PoEWSPn/OZDRiyXlyMDG1wTP7ASeLajaaPajbuv
YorweKeY8mKggHeAudUbjE4rE2zW8fge8gKugUAIUgEYBGyri+ASKrWlb6ym2QRL
8u+qiGSQwit2eBeY0i390ahc+oe9rgPC1y0Uc4ZlkY+Cc9X/Tuhb75Y8ObiXhoCh
xPbMLaKXTapXTxteOjJYWKZ4F4hiUCYyol+yGqSP3ux0yPX7wcLhtOlswNubNKly
J9citm5ZBMTvydD6m6JDWbULAglxNSXEINLNuxLtNoFIM0s8ECzpkwCfPg==
-----END CERTIFICATE-----`)

var keyPem = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAxPaFgXrchLFs5mZ3sZj4e4Ba/gOlOf1nNBAjNJ9hnXoKbIIL
T+5+GT349LZB57iZj9+FM2+Bs2Y57kJ/957oLPfzlC/PXsveiutoWV3zeY7+binq
LMYEWgmP+wjunZ8tDj4k0EJr2PsHZlx73ST+dOUVVwvOQl7c2vgrHN8pSc+5z9Qk
8RWGvBQm4Trk1Ciud7v6nc0Uh45WQwLdMJqRycsK+MBUHpe8TXkbCh4quscMw34p
xgWAg/hfenLot2U9qLW5zrhgCBP95h4hw0aulZXBB/mxADiyYElKLsBF7mgnC7ZI
QmxmycbiaflWLwD/Iei+9gKfVMI/8kzNKdf4LQIDAQABAoIBADYo55ssEpk2RJCy
WnVub91d9SdmDzf78zYAvf2JWgk4dsdRlxS6qtf8D4oS19qFC0zhlLoJDmwrTwCy
LogDnSpIYCU+ZFJX0vD2PHJegEXLyTC3u9nl9WpguMO5uAuFqpkBA5R0vz0iAe3m
vnSX6JRyLcUKzQO5HBfmJ8y/nJXb2eB7pF9Iwlflck2h5lQdSaN6UUBNZ/L9MMh/
CW68CMRpHEJFu/oGi4wyEt8zk3REH6m46CYv0LxI/Au/5jRtN8RnN/L3h8xxqyNm
rW9btAnREkjQEL5LLWPbHjHevtWB6fc4El6yaL8sWS0n9l0xTtl0XGDVKxNtB1Xh
O+fPWe0CgYEA9mxJ0UtjbddGSRHW6CXv27Qrg5T18Lo0sBTq3ouUqhIUXd1SAejf
iLG2j1URzJY3C9HanX0lRA68SMJhQDlvye89/b14hxmDwxk874LxPX8VJvxbR6FV
qPg/bnr8vBmE9BKRvlmv8f20jB+y7hQmYJL9fC2Y5BoAxr9C9lOJbWcCgYEAzJ4l
PczNTmzwnFIqNpUj5uHogFD0AI1k9cdcbKDE2bC55DDz3WoWnk+NUPLhD0LXO1hz
hUBlSZF+ZWElOFHQ2wTmYJ9711daEcmhEwbbEZNwgVuMjBWLUkRGkosUZrKKtJVr
JYkJLLWcdCEIt+VU7e25ePI9qq+J9g3aqlDl3UsCgYEA3zsf7+qUawfOUxlHDsx3
KhdgJ/YEiguU+UIptmrJxPtV2eZJiRNVlHYxBE0zL5uQyDNWEL8yyCF1LZBxGwYt
H8iL5tYCXpidhVrSmcKMGYKLPeL0KcKcX9JrXAEr/JY9nAFKaB7FRbnoGdwJcqVs
UqY13Y1M6K4pr+HJnm30m9cCgYAUzyyIaXCjvi6GJ1EFtgstquHjUthNyhNvb3P1
1C/Q18k/7L6QUP614O9FQT4kOC79aRRug8sJPVO2abfIT4HHFGt9fhqxHsAZOQE5
lyPmWLFDZpUXlgVSO4FV2/EaNKQok12PNq2JL3sW0Fk7ooYNoHSRWUluN2X3cRdA
5PNLmwKBgQDJRj4llNNLFvhcnbz/qc+8WbIAM99CSDLWShYwNr/mv/DhsZJFtJw1
H3toso6NwcaT15Stx85c/6BT7sfiLW072yh4T6+p85VEEHjLePndM8dkesckBXPs
jEtbNcmlkfwjidFUpS3y9ADIIPm2LLr0oQXil+wSY9HDiYrtIkJAWQ==
-----END RSA PRIVATE KEY-----`)

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
	cert, err := tls.X509KeyPair(certPem, keyPem)
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
