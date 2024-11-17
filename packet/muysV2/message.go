package muysV2

// Format: |--Length(2)--|--Data(variable)--|
type Message struct {
	length  int32  // 4字节 = length(4) +  msgType（2长度）+ data 长度  本来只放 data 长度，放 msgType 长度方便解包计算
	msgType uint16 // 2字节 消息类型：1:系统 2:login 3:game etc
	data    []byte // Payload data [msgId, protoMessage]
}

func (that *Message) Name() string {
	return Name
}
func (that *Message) GetData() []byte {
	return that.data
}

func NewMessage(msgType uint16, data []byte) *Message {
	le := defaultSizeBytes + len(data)
	return &Message{
		length:  int32(le) + defaultTypeBytes, // 放 msgType 长度方便解包计算
		msgType: msgType,
		data:    data,
	}
}
