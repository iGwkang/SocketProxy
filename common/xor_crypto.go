package common

import (
	"net"
	"unsafe"
)

const uint64Size = int(unsafe.Sizeof(uint64(0)))

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
	dw := *(*[]uint64)(unsafe.Pointer(&dst))
	var uintB uint64
	for i := 0; i < uint64Size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&uintB)) + uintptr(i))) = b
	}

	n := len(dst) / uint64Size

	nx := n % 8
	for i := 0; i < nx; i++ {
		dw[i] ^= uintB
	}

	for i := nx; i < n; i += 8 {
		_dst := dw[i : i+8]
		_dst[0] ^= uintB
		_dst[1] ^= uintB
		_dst[2] ^= uintB
		_dst[3] ^= uintB

		_dst[4] ^= uintB
		_dst[5] ^= uintB
		_dst[6] ^= uintB
		_dst[7] ^= uintB
	}

	ex := len(dst) % 8
	for i := len(dst) - ex; i < len(dst); i++ {
		dst[i] ^= b
	}
}
