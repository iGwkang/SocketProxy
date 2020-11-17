package main

import (
	. "SocketProxy/logger"
	"bufio"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/patrickmn/go-cache"
)

// 走代理的域名
var proxy_list []*regexp.Regexp

// 不走代理的域名
var bypass_list []*regexp.Regexp

// 访问过的域名缓存
var domainCache = cache.New(1*time.Hour, 0)

func InitDomainList() (err error) {
	var reader io.ReadCloser
	reader, err = os.Open(ClientConfig.DomainFile)
	if err != nil {
		// 如果没有文件 直接访问github获取
		var resp *http.Response
		resp, err = http.Get("https://raw.githubusercontent.com/NateScarlet/gfwlist.acl/master/gfwlist.acl")
		if err != nil {
			return err
		}
		reader = resp.Body
	}
	defer reader.Close()

	const (
		PROXY_LIST  = 1
		BYPASS_LIST = 2
	)

	var listValue = 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "[proxy_list]" {
			listValue = PROXY_LIST
		} else if text == "[bypass_list]" {
			listValue = BYPASS_LIST
		} else if text != "" && listValue != 0 {
			compile, err := regexp.Compile(scanner.Text())
			if err == nil {
				switch listValue {
				case PROXY_LIST:
					proxy_list = append(proxy_list, compile)
				case BYPASS_LIST:
					bypass_list = append(bypass_list, compile)
				}
			}
		}
	}
	proxy_list = proxy_list[:len(proxy_list):len(proxy_list)]
	bypass_list = proxy_list[:len(bypass_list):len(bypass_list)]
	Logger.Info("init DomainList. [proxy_list] Num =", len(proxy_list), " [bypass_list] Num =", len(bypass_list))
	return
}

func DomainIsProxy(domain string) bool {
	// 先判断缓存里有没有
	item, ok := domainCache.Get(domain)
	if ok {
		return item.(bool)
	}

	// 不走代理的列表
	for i := 0; i < len(bypass_list); i++ {
		if bypass_list[i].MatchString(domain) {
			domainCache.Set(domain, false, 0)
			return false
		}
	}

	// 走代理的列表
	for i := 0; i < len(proxy_list); i++ {
		if proxy_list[i].MatchString(domain) {
			domainCache.Set(domain, true, 0)
			return true
		}
	}
	// 都没找到, 不走代理
	return false
}
