package packet_test

import (
	"github.com/cute-angelia/go-game-utils/encoding/proto"
	"github.com/cute-angelia/go-game-utils/packet/muys"
	"github.com/cute-angelia/go-game-utils/packet/qx"
	"testing"
)

func TestQx(t *testing.T) {
	// 1. []byte
	var packer = qx.NewPacker(qx.WithCodeC(""))
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
	t.Logf("name: %s", message.Name())
	t.Logf("mainid: %d", msgdecode.GetMainID())
	t.Logf("subid: %d", msgdecode.GetSubID())
	t.Logf("data: %v", msgdecode.GetData())
	t.Log(string(msgdecode.GetData()))
}

func TestQx2(t *testing.T) {
	// 2. proto
	var packer = qx.NewPacker(qx.WithCodeC("proto"))

	// 示例 pb 随意写的
	h := qx.TestData{Code: 1, Msg: "测试 pb encoding", Name: "test"}

	msg := qx.NewMessage(311, 2, &h)
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
	t.Logf("name: %s", message.Name())
	t.Logf("mainid: %d", msgdecode.GetMainID())
	t.Logf("subid: %d", msgdecode.GetSubID())
	t.Logf("data: %v", msgdecode.GetData())

	// 解析 pb
	hresp := qx.TestData{}
	proto.Unmarshal(msgdecode.GetData(), &hresp)
	t.Logf("%+v", &hresp)
}

func TestQx3(t *testing.T) {
	// 3. proto and client
	var packer = qx.NewPacker(qx.WithCodeC("proto"), qx.WithIsClient(true))

	// 示例 pb 随意写的
	h := qx.TestData{Code: 1, Msg: "测试 pb encoding", Name: "test"}

	msg := qx.NewMessage(311, 2, &h)
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
	t.Logf("name: %s", message.Name())
	t.Logf("mainid: %d", msgdecode.GetMainID())
	t.Logf("subid: %d", msgdecode.GetSubID())
	t.Logf("data: %v", msgdecode.GetData())

	// 解析 pb
	hresp := qx.TestData{}
	qx.UnmarshalPb(msgdecode.GetData(), &hresp)
	t.Logf("%+v", &hresp)
}

func TestDefaultPackerMuys(t *testing.T) {
	var packer3 = muys.NewPacker(muys.WithEndian(muys.BigEndian))
	msg := muys.NewMessage([]byte("hello muys"))
	data, err := packer3.PackMessage(msg)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)

	message, err := packer3.UnpackMessage(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("name: %s", message.Name())
}
