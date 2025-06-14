package util

import (
	"bytes"
	bin "encoding/binary"
	"fmt"
	"io"
)

func WriteMarker(b *bytes.Buffer) {
	b.Write([]byte{0xFF})
}

func ReadMarker(r io.Reader) {
	buf := make([]byte, 1)
	r.Read(buf)
	if buf[0] != 0xFF {
		panic(fmt.Sprintf("Expected marker, got %d", buf[0]))
	}
}

func ReadString(r io.Reader) string {
	var len uint32
	bin.Read(r, bin.LittleEndian, &len)
	buf := make([]byte, len)
	r.Read(buf)
	return string(buf)
}

func WriteString(b *bytes.Buffer, s string) {
	bin.Write(b, bin.LittleEndian, uint32(len(s)))
	b.Write([]byte(s))
}
