package common

import (
	"net"
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
		return 0, err
	}
	for i := 0; i < len(outData); i++ {
		outData[i] ^= c.xorByte
	}
	return n, nil
}

func (c *xorCipher) Write(data []byte) (int, error) {
	for i := 0; i < len(data); i++ {
		data[i] ^= c.xorByte
	}
	return c.Conn.Write(data)
}
