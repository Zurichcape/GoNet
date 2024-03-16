package interfaces

type IMessage interface {
	// GetMsgId 获取消息ID
	GetMsgId() uint32
	// GetMsgLen 获取消息长度
	GetMsgLen() uint32
	// GetData 获取消息内容
	GetData() []byte

	// SetMsgId 设置消息ID
	SetMsgId(uint32)
	// SetMsgLen 设置消息长度
	SetMsgLen(uint32)
	// SetMsgData 设置消息
	SetMsgData([]byte)
}
