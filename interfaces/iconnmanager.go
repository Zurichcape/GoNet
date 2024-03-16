package interfaces

type IConnMgr interface {
	// AddConn 添加连接
	AddConn(IConnection)
	// DeleteConn 删除连接
	DeleteConn(IConnection)
	// GetConn 根据ConnID返回连接
	GetConn(connID uint64) (IConnection, error)
	// GetConnLen 得到当前连接数
	GetConnLen() int
	// ClearConn 清除所有连接
	ClearConn()
}
