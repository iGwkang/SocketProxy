package common

import (
	"crypto/sha1"
	"encoding/binary"
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

func TestStringToBytes(t *testing.T) {

}

func BenchmarkStringToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		passwd := "afafasdajgdhvcziydawgajsdhgjahsgdhcfjhzxc"
		buf := make([]byte, len(passwd) + 6)
		copy(buf, StringToBytes(passwd))
		binary.BigEndian.PutUint32(buf[len(passwd):], Inet_addr("127.0.0.2"))
		binary.BigEndian.PutUint16(buf[len(passwd) + 4:], 865)

		//dstAddr := make([]byte, 6)
		//binary.BigEndian.PutUint32(dstAddr, Inet_addr("127.0.0.2"))
		//binary.BigEndian.PutUint16(dstAddr[4:6], 865)
		//
		//buf := []byte(passwd)
		//buf = append(buf, dstAddr...)

	}
}
