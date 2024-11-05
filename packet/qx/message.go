package qx

import "google.golang.org/protobuf/proto"

// Format: |--Length(4)--|--MainID(4)--|--SubID(4)--|--Data(variable)--|
type Message struct {
	length int32  // 总长度 = 【 12 (3个int32) + 包长 】
	mainID int32  // Main ID of the packet
	subID  int32  // Sub ID of the packet
	data   []byte // Payload data
}

const MaxSubId = 10000000

func (that *Message) Type() string {
	return "qx"
}
func (that *Message) Route() int32 {
	return EncodeRoute(that.mainID, that.subID)
}

func (that *Message) Data() []byte {
	return that.data
}

func (that *Message) GetMainID() int32 {
	return that.mainID
}
func (that *Message) GetSubID() int32 {
	return that.subID
}

func NewMessage(mainId, subId int32, pb proto.Message) *Message {
	data, _ := proto.Marshal(pb)
	le := defaultSizeBytes + defaultMainIdBytes + defaultSubIdBytes + len(data)
	return &Message{
		length: int32(le),
		mainID: mainId,
		subID:  subId,
		data:   data,
	}
}

func NewMessageByte(mainId, subId int32, data []byte) *Message {
	le := defaultSizeBytes + defaultMainIdBytes + defaultSubIdBytes + len(data)
	return &Message{
		length: int32(le),
		mainID: mainId,
		subID:  subId,
		data:   data,
	}
}

// EncodeRoute 组合两个ID成为一个数字
func EncodeRoute(mainID, subID int32) int32 {
	return mainID*MaxSubId + subID
}

// DecodeRoute 解码组合后的数字，返回 mainID 和 subID
func DecodeRoute(combinedID int32) (mainID, subID int32) {
	mainID = combinedID / MaxSubId
	subID = combinedID % MaxSubId
	return
}
