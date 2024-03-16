package interfaces

type IRequest interface {
	// GetConn  得到当前连接
	GetConn() IConnection
	// GetData 得到请求的数据
	GetData() []byte
	// GetMsgID 得到请求数据的ID
	GetMsgID() uint32
}
