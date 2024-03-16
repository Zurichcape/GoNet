package net

import (
	"errors"
	"fmt"
	"gonet/interfaces"
	"sync"
)

var _ interfaces.IConnMgr = (*ConnManager)(nil)

type ConnManager struct {

	//管理连接集合
	connections map[uint64]interfaces.IConnection
	//保护连接集合的读写锁
	connLock sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint64]interfaces.IConnection),
	}
}

// AddConn  添加连接
func (cm *ConnManager) AddConn(conn interfaces.IConnection) {
	//保护共享资源,加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	//将conn加入connManager中
	cm.connections[conn.GetConnID()] = conn
	fmt.Println("connection add to connManager successfully: conn num= ", cm.GetConnLen())
}

// DeleteConn  删除连接
func (cm *ConnManager) DeleteConn(conn interfaces.IConnection) {
	//保护共享资源,加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	delete(cm.connections, conn.GetConnID())
	fmt.Println("ConnID = ", conn.GetConnID(), " ,delete from connManager successfully: conn num= ", cm.GetConnLen())
}

// GetConn  根据ConnID返回连接
func (cm *ConnManager) GetConn(connID uint64) (conn interfaces.IConnection, err error) {
	//保护共享资源,加读锁
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if connection, ok := cm.connections[connID]; ok {
		return connection, nil
	}
	return nil, errors.New("conn doesn't exit")
}

// GetConnLen 得到当前连接数
func (cm *ConnManager) GetConnLen() int {
	return len(cm.connections)
}

// ClearConn  清除所有连接
func (cm *ConnManager) ClearConn() {
	//保护共享资源,加写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	//删除conn并停止conn的工作
	for connID, conn := range cm.connections {
		//停止
		conn.Stop()
		//删除
		delete(cm.connections, connID)
	}
}
