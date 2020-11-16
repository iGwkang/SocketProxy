package main

import (
	"SocketProxy/common"
	. "SocketProxy/logger"
	"bufio"
	"os"
	"strings"
)

var whiteIPRange []common.IPRange

// 初始化白名单列表
func InitWhiteIPList() error {
	file, err := os.Open(ClientConfig.IPWhiteFile)
	if err != nil {
		Logger.Warn(err)
		return err
	}
	defer file.Close()

	// 读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), "-")
		if len(split) == 2 {
			whiteIPRange = append(whiteIPRange, common.IPRange{
				Begin: common.Inet_addr_h(split[0]),
				End:   common.Inet_addr_h(split[1]),
			})
		}
	}

	// TODO: sort

	Logger.Info("Init WhiteIP List. [white ip list] Num =", len(whiteIPRange))
	return nil
}

// 判断ip是否在白名单
func IPisProxy(ip string) bool {
	if common.IsIntranetAddress(ip) {
		return false
	}

	ipUint := common.Inet_addr_h(ip)
	if ipUint == 0 {
		return false
	}

	index := 0
	left := 0
	right := len(whiteIPRange) - 1

	for left <= right {
		index = left + (right-left)/2

		if ipUint < whiteIPRange[index].Begin {
			right = index - 1
		} else if ipUint > whiteIPRange[index].End {
			left = index + 1
		} else {
			return false
		}
	}

	return true
}
