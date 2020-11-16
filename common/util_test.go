package common

import (
	"testing"
)

func TestInet_addr_h(t *testing.T) {
	t.Log(Inet_addr_h("127.0.0.1"))
	t.Log(Inet_addr_h("127.0.0.2"))
}

func TestIsIntranetAddress(t *testing.T) {
	t.Log(IsIntranetAddress("35.65.32.22"))
	t.Log(IsIntranetAddress("127.0.0.2"))
}
