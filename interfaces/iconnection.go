package interfaces

import "net"

type IConnection interface {
	// Start 启动连接，让当前连接开始工作
	Start()

	// Stop 停止链接 结束当前连接的工作
	Stop()

	// GetTCPConnection 获取当前连接绑定的conn套接字
	GetTCPConnection() *net.TCPConn

	// GetConnID 获取当前连接模块的连接ID
	GetConnID() uint64

	// RemoteAddr 获取远程客户端的TCP状态 Host port
	RemoteAddr() net.Addr

	// SendMsg 发送数据，将数据发送给远程的客户端
	SendMsg(uint32, []byte) error

	// SetProperty 设置连接属性
	SetProperty(string, interface{})

	// GetProperty 获取连接属性
	GetProperty(string) (interface{}, error)

	// DeleteProperty 删除连接属性
	DeleteProperty(string)
}
