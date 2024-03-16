package net

import "gonet/interfaces"

var _ interfaces.IRequest = (*Request)(nil)

type Request struct {
	//已经和客户端建立的连接
	conn interfaces.IConnection
	//客户端请求的数据
	//data []byte
	msg interfaces.IMessage
}

// GetConn 得到当前连接
func (r *Request) GetConn() interfaces.IConnection {
	return r.conn
}

// GetData 得到当前数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}
func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgId()
}
