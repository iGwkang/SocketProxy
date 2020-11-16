package main

func main() {
	go func() {
		dnsServer := NewDNSServer()
		dnsServer.Run()
	}()

	tcpServer := NewTcpServer()
	tcpServer.Run()
}
