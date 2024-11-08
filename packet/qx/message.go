package qx

import "google.golang.org/protobuf/proto"

const MaxSubId = 10000000

// Format: |--Length(4)--|--MainID(4)--|--SubID(4)--|--Data(variable)--|
type Message struct {
	length int32 // 总长度 = 【 12 (3个int32) + 包长 】
	mainID int32 // Main ID of the packet
	subID  int32 // Sub ID of the packet
	data   any   // Payload data 根据 codec 变化
}

func (that *Message) Name() string {
	return Name
}

// data 需传入指针
func NewMessage(mainId, subId int32, data any) *Message {
	return &Message{
		length: 0, // 获取 code 后设置
		mainID: mainId,
		subID:  subId,
		data:   data,
	}
}

// GetData 返回数据
func (that *Message) GetData() []byte {
	return that.data.([]byte)
}

// ==================== 特殊方法 ======================

func (that *Message) GetMainID() int32 {
	return that.mainID
}
func (that *Message) GetSubID() int32 {
	return that.subID
}

func (that *Message) UnmarshalPb(v proto.Message) error {
	return UnmarshalPb(that.GetData(), v)
}

func UnmarshalPb[T proto.Message](buf []byte, v T) error {
	return proto.Unmarshal(buf, v)
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
