package ipacket

type Packer interface {
	// ReadMessage 读取消息
	ReadMessage(reader interface{}) ([]byte, error)
	// PackMessage 打包消息
	PackMessage(message Message) ([]byte, error)
	// UnpackMessage 解包消息
	UnpackMessage(data []byte) (Message, error)
	// PackHeartbeat 打包心跳
	PackHeartbeat() ([]byte, error)
	// CheckHeartbeat 检测心跳包
	CheckHeartbeat(data []byte) (bool, error)
	// UnmarshalData 反解析
	UnmarshalData(data []byte, v interface{}) error
	// String get name
	String() string
}

type Message interface {
	Name() string // 类型
	GetData() []byte
}

//func NewMessageQx(mainId, subId int32, data []byte) Message {
//	return qx.NewMessage(mainId, subId, data)
//}
//
//func NewPackerQx(endian string, bufferBytes int) *qx.Packer {
//	return qx.NewPacker(qx.WithBufferBytes(bufferBytes), qx.WithEndian(endian))
//}
