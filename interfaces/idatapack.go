package interfaces

/*
	封包、拆包模块
	直接面向TCP连接中的数据流，用于处理TCP粘包问题
*/

type IDataPack interface {
	// GetHeadLen 获取包的头部长度
	GetHeadLen() uint32
	// Pack 封包方法
	Pack(IMessage) ([]byte, error)
	// UnPack 拆包方法
	UnPack([]byte) (IMessage, error)
}

const (
	GoNetDataPack string = "gonet_pack"

	//...(+)
	//自定义封包方式在此添加
)

const (
	GoNetMessage string = "gonet_message"
	//...(+)
	//自定义消息方式
)
