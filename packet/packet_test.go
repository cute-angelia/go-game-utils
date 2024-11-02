package packet_test

import (
	"go-game-utils/packet"
	"go-game-utils/packet/qx"
	"testing"
)

var packer = qx.NewPacker(qx.WithEndian(packet.LittleEndian))

func TestDefaultPacker_PackMessage(t *testing.T) {
	msg := qx.NewMessage(311, 2, []byte("hello world"))
	data, err := packer.PackMessage(msg)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)

	message, err := packer.UnpackMessage(data)
	if err != nil {
		t.Fatal(err)
	}

	msgdecode := message.(*qx.Message)

	t.Logf("type: %s", message.Type())
	t.Logf("route: %d", message.Route())
	t.Logf("buffer: %s", string(message.Data()))
	t.Logf("mainid: %d", msgdecode.GetMainID())
	t.Logf("subid: %d", msgdecode.GetSubID())
}
