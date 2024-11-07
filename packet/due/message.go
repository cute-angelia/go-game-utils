package due

type Message struct {
	Seq    int32  // 序列号
	Route  int32  // 路由ID
	Buffer []byte // 消息内容
}

func (that *Message) Name() string {
	return "muys"
}
func (that *Message) GetData() []byte {
	return that.Buffer
}
