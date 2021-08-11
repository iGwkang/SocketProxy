package common

import (
	"crypto/sha1"
	"net/http"
	"strings"
	"testing"
)

func TestInet_addr_h(t *testing.T) {
	t.Log(Inet_addr("127.0.0.1"))
	t.Log(Inet_addr("127.0.0.2"))

	str := "127.0.0.1:2345"
	str = str[:strings.IndexByte(str, ':')]

	t.Log(str)
}

func TestIsIntranetAddress(t *testing.T) {
	//t.Log(IsIntranetAddress("35.65.32.22"))
	// t.Log(IsIntranetAddress("127.0.0.2"))
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 512)

	n, err := resp.Body.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf[:n]))
}

func TestSha1Sum(t *testing.T) {
	buf := sha1.Sum([]byte("test"))
	str := string(buf[:])
	t.Log(buf)
	t.Log(len(str))
}
