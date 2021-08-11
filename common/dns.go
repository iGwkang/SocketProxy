package common

import (
	"net"
	"time"
)

// dns请求
func RequestDNSParse(dnsMsg []byte, dnsServer string, timeout time.Duration) (data []byte, err error) {
	addr, err := net.ResolveUDPAddr("udp", dnsServer)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()
	if timeout > 0 {
		err = conn.SetDeadline(time.Now().Add(timeout))
		if err != nil {
			return
		}
	}

	_, err = conn.Write(dnsMsg)
	if err != nil {
		return
	}

	readMsg := make([]byte, 1024)

	n, err := conn.Read(readMsg)
	if err != nil {
		return
	}
	data = readMsg[:n]

	return
}
