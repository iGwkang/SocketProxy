package common

import (
	"crypto/rand"
	"io"
	"net"
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
