package muys

// Format: |--Length(2)--|--Data(variable)--|
type Message struct {
	length uint16 // 2字节 内容：仅仅消息长度
	data   []byte // Payload data
}

func (that *Message) Name() string {
	return Name
}
func (that *Message) GetData() []byte {
	return that.data
}

func NewMessage(data []byte) *Message {
	le := defaultSizeBytes + len(data)
	return &Message{
		length: uint16(le),
		data:   data,
	}
}
