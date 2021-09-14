package main

func main() {
	go func() {
		dnsClient := NewDNSClient(conf.ListenDNSAddr, conf.DNSServer, conf.Timeout)
		dnsClient.Run()
	}()

	go func() {
		httpServer := NewLocalTCPServer(conf.ListenTcpAddrs, conf.TcpServerAddrs, conf.Password, conf.Timeout)
		httpServer.Run()
	}()

	select {}
}
