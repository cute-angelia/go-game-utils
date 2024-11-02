package muys

// Format: |--Length(2)--|--Data(variable)--|
type Message struct {
	length int32  // 2字节 内容：仅仅消息长度
	data   []byte // Payload data
}

func (that *Message) Type() string {
	return "muys"
}
func (that *Message) Route() int32 {
	return 0
}

func (that *Message) Data() []byte {
	return that.data
}

func NewMessage(data []byte) *Message {
	le := defaultSizeBytes + len(data)
	return &Message{
		length: int32(le),
		data:   data,
	}
}
