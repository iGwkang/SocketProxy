package common

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

// (for)  BenchmarkXor-6   	 2878296	       6402 ns/op	       0 B/op	       0 allocs/op
// (safe) BenchmarkXor-6   	 4015207	       5588 ns/op	       0 B/op	       0 allocs/op
// (fast) BenchmarkXor-6   	 1486964	       802 ns/op	       0 B/op	       0 allocs/op
func BenchmarkXor(b *testing.B) {
	buf := make([]byte, 16000)
	_, _ = io.ReadFull(rand.Reader, buf)

	buf1 := make([]byte, 16000)
	copy(buf1, buf)
	b.ResetTimer()

	for j := 0; j < b.N; j++ {
		// for i := 0; i < len(buf); i++ {
		// 	buf[i] ^= 22
		// }
		//SafeXORBytes(buf, 22)
		FastXORByte(buf, 22)
	}
}

func TestXor(t *testing.T) {
	buf := make([]byte, 16003)
	_, _ = io.ReadFull(rand.Reader, buf)

	buf1 := make([]byte, 16003)
	copy(buf1, buf)

	t.Log(bytes.Equal(buf, buf1))

	SafeXORBytes(buf, 22)
	FastXORByte(buf1, 22)

	t.Log(bytes.Equal(buf, buf1))
}
