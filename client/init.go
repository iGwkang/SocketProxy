package main

import (
	"SocketProxy/logger"
	"flag"
)

func init() {
	flag.Parse()
	logger.InitLogger() // 初始化日志
	LoadConfig()        // 读取配置文件
	InitTLSConfig()     // 初始化tls配置
	InitDomainList()	// 初始化域名列表
	InitWhiteIPList()	// 初始化ip白名单
}