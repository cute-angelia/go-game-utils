package buffer_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go-game-utils/buffer"
	"testing"
)

type User struct {
	ID  int32
	Age int8
}

func TestNewBuffer(t *testing.T) {
	buff := &bytes.Buffer{}
	buff.Grow(2)

	binary.Write(buff, binary.BigEndian, int16(2))

	fmt.Println(buff.Bytes())

	writer := buffer.NewWriter(2)
	writer.WriteInt16s(binary.BigEndian, int16(2))

	fmt.Println(writer.Bytes())

	writer.Reset()
	writer.WriteInt16s(binary.BigEndian, int16(20))
	writer.WriteFloat32s(binary.BigEndian, 5.2)

	fmt.Println(writer.Bytes())

	data := writer.Bytes()

	reader := buffer.NewReader(data)
	v1, _ := reader.ReadInt16(binary.BigEndian)
	fmt.Println(v1)
	v2, _ := reader.ReadFloat32(binary.BigEndian)
	fmt.Println(v2)
}
