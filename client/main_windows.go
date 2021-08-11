package main

func main() {
	go func() {
		dnsClient := NewDNSClient(conf.ListenDNSAddr, conf.DNSServer, conf.Timeout)
		dnsClient.Run()
	}()

	go func() {
		httpServer := NewLocalHttpServer(conf.ListenHttpAddr, conf.TcpServerAddrs, conf.Encryption, conf.Password, conf.Timeout)
		httpServer.Run()
	}()

	go func() {
		socksServer := NewLocalSocks5Server(conf.ListenSocks5Addr, conf.TcpServerAddrs, conf.Encryption, conf.Password, conf.Timeout)
		socksServer.Run()
	}()
	select {}
}
