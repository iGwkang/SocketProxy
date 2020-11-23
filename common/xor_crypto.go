package common

import (
	"net"
	"unsafe"
)

// 异或加解密
type xorCipher struct {
	net.Conn
	xorByte byte
}

func NewXorCipher(xorByte byte, conn net.Conn) net.Conn {
	return &xorCipher{
		Conn:    conn,
		xorByte: xorByte,
	}
}

func (c *xorCipher) Read(outData []byte) (int, error) {
	n, err := c.Conn.Read(outData)
	if err != nil {
		return n, err
	}
	FastXORByte(outData, c.xorByte)
	return n, nil
}

func (c *xorCipher) Write(data []byte) (int, error) {
	FastXORByte(data, c.xorByte)
	return c.Conn.Write(data)
}

func SafeXORBytes(dst []byte, b byte) {
	ex := len(dst) % 8
	for i := 0; i < ex; i++ {
		dst[i] ^= b
	}

	for i := ex; i < len(dst); i += 8 {
		_dst := dst[i : i+8]
		_dst[0] ^= b
		_dst[1] ^= b
		_dst[2] ^= b
		_dst[3] ^= b

		_dst[4] ^= b
		_dst[5] ^= b
		_dst[6] ^= b
		_dst[7] ^= b
	}
}

func FastXORByte(dst []byte, b byte) {
	const uint64Size = 8

	uint64Array := *(*[]uint64)(unsafe.Pointer(&dst))
	bByte := uint64(b) +
		uint64(b)<<(8*1) +
		uint64(b)<<(8*2) +
		uint64(b)<<(8*3) +
		uint64(b)<<(8*4) +
		uint64(b)<<(8*5) +
		uint64(b)<<(8*6) +
		uint64(b)<<(8*7)

	// for i := 0; i < uint64Size; i++ {
	// 	*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&uintB)) + uintptr(i))) = b
	// }

	n := len(dst) / uint64Size

	nx := n % 8
	for i := 0; i < nx; i++ {
		uint64Array[i] ^= bByte
	}

	for i := nx; i < n; i += 8 {
		tmpArray := uint64Array[i : i+8]
		tmpArray[0] ^= bByte
		tmpArray[1] ^= bByte
		tmpArray[2] ^= bByte
		tmpArray[3] ^= bByte

		tmpArray[4] ^= bByte
		tmpArray[5] ^= bByte
		tmpArray[6] ^= bByte
		tmpArray[7] ^= bByte
	}

	ex := len(dst) % 8
	for i := len(dst) - ex; i < len(dst); i++ {
		dst[i] ^= b
	}
}
