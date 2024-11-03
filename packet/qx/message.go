package qx

// Format: |--Length(4)--|--MainID(4)--|--SubID(4)--|--Data(variable)--|
type Message struct {
	length int32  // 总长度 = 【 12 (3个int32) + 包长 】
	mainID int32  // Main ID of the packet
	subID  int32  // Sub ID of the packet
	data   []byte // Payload data
}

func (that *Message) Type() string {
	return "qx"
}
func (that *Message) Route() int32 {
	return that.mainID*10000000 + that.subID
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

func NewMessage(mainId, subId int32, data []byte) *Message {
	le := defaultSizeBytes + defaultMainIdBytes + defaultSubIdBytes + len(data)
	return &Message{
		length: int32(le),
		mainID: mainId,
		subID:  subId,
		data:   data,
	}
}
