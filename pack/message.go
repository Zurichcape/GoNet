package pack

type Message struct {
	ID      uint32 //消息ID
	DataLen uint32 //消息长度
	Data    []byte //消息内容
}

// NewMessage 创建一个Message消息包
func NewMessage(id uint32, data []byte) *Message {
	return &Message{
		ID:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// GetMsgId 获取消息ID
func (m *Message) GetMsgId() uint32 {
	return m.ID
}

// GetMsgLen 获取消息长度

func (m *Message) GetMsgLen() uint32 {
	return m.DataLen
}

// GetData 获取消息内容

func (m *Message) GetData() []byte {
	return m.Data
}

// SetMsgId 设置消息ID
func (m *Message) SetMsgId(id uint32) {
	m.ID = id
}

// SetMsgLen 设置消息长度
func (m *Message) SetMsgLen(len uint32) {
	m.DataLen = len
}

// SetMsgData 设置消息
func (m *Message) SetMsgData(data []byte) {
	m.Data = data
}
