package packet

import (
	"github.com/cute-angelia/go-game-utils/packet/due"
	"github.com/cute-angelia/go-game-utils/packet/ipacket"
	"github.com/cute-angelia/go-game-utils/packet/muys"
	"github.com/cute-angelia/go-game-utils/packet/qx"
)

func GetDefaultPacker(packerName string) ipacket.Packer {
	switch packerName {
	case "due":
		return due.NewPacker()
	case "muys":
		return muys.NewPacker()
	case "qx":
		return qx.NewPacker(qx.WithCodeC("proto"))
	}
	return nil
}
