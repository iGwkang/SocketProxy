package common

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// 获取非零数
func GetNonZeroNumber() uint8 {
	n := [1]byte{}
	for {
		rand.Read(n[:])
		if n[0] != 0 {
			return n[0]
		}
	}
}

// 地址转为字符串
func AddrToString(addr []byte) (ip, port string) {
	if len(addr) != 6 {
		return
	}

	ip += strconv.FormatInt(int64(addr[0]), 10) + "."
	ip += strconv.FormatInt(int64(addr[1]), 10) + "."
	ip += strconv.FormatInt(int64(addr[2]), 10) + "."
	ip += strconv.FormatInt(int64(addr[3]), 10)
	port = strconv.FormatInt(int64(binary.BigEndian.Uint16(addr[4:6])), 10)
	return
}

// 获取外网ip
func GetExternalIP() net.IP {
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	buf := make([]byte, 32)

	n, err := resp.Body.Read(buf)
	if err != nil {
		return nil
	}
	ip := string(buf[:n])
	return net.ParseIP(ip).To4()
}

// 交换两个连接的数据
func Relay(left, right net.Conn) {
	ch := make(chan struct{})
	defer close(ch)
	go func() {
		_, _ = io.Copy(right, left)
		_ = right.SetDeadline(time.Now())
		_ = left.SetDeadline(time.Now())
		ch <- struct{}{}
	}()

	_, _ = io.Copy(left, right)
	_ = right.SetDeadline(time.Now())
	_ = left.SetDeadline(time.Now())
	<-ch
}
