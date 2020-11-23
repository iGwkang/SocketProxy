package common

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
	"unsafe"
)

// (for)  BenchmarkXor-6   	 2878296	       6402 ns/op	       0 B/op	       0 allocs/op
// (safe) BenchmarkXor-6   	 4015207	       5588 ns/op	       0 B/op	       0 allocs/op
// (fast) BenchmarkXor-6   	 1486964	       679 ns/op	       0 B/op	       0 allocs/op
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

	var bByte uint64
	tNum := GetNonZeroNumber()
	for i := 0; i < int(unsafe.Sizeof(bByte)); i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&bByte)) + uintptr(i))) = tNum
	}
	t.Log(bByte)
	bByte = uint64(tNum) +
		uint64(tNum)<<(8*1) +
		uint64(tNum)<<(8*2) +
		uint64(tNum)<<(8*3) +
		uint64(tNum)<<(8*4) +
		uint64(tNum)<<(8*5) +
		uint64(tNum)<<(8*6) +
		uint64(tNum)<<(8*7)
	t.Log(bByte)
}

func BenchmarkByte(b *testing.B) {
	tNum := GetNonZeroNumber()
	var bByte uint64
	for j := 0; j < b.N; j++ {
		// for i := 0; i < int(unsafe.Sizeof(bByte)); i++ {
		// 	*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&bByte)) + uintptr(i))) = tNum
		// }

		bByte = uint64(tNum) +
			uint64(tNum)<<(8*1) +
			uint64(tNum)<<(8*2) +
			uint64(tNum)<<(8*3) +
			uint64(tNum)<<(8*4) +
			uint64(tNum)<<(8*5) +
			uint64(tNum)<<(8*6) +
			uint64(tNum)<<(8*7)

	}
	_ = bByte
}
