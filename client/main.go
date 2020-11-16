package main

func main() {
	go func() {
		dnsClient := NewDNSClient()
		dnsClient.Run()
	}()

	tcpClient := NewTcpClient()
	tcpClient.Run()
}
