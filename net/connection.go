package net

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gonet/config"
	"gonet/interfaces"
	"gonet/pack"
	"io"
	"net"
	"sync"
)

/*
连接模块
*/
var _ interfaces.IConnection = (*Connection)(nil)

type Connection struct {
	//当前connection属于哪个server
	TcpServer interfaces.IServer

	//当前连接的TCP socket套接字
	Conn *net.TCPConn

	//连接的ID, 也可以称作为SessionID，ID全局唯一
	ConnID uint64

	//当前的连接状态
	isClosed bool

	//告知当前连接已经退出/停止的channel(由Reader告知Writer停止)
	//ExitBuffChan chan bool
	// 告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//无缓冲管道，用于读、写Goroutine之间的消息通信
	msgChan chan []byte
	sync.RWMutex

	// 消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler interfaces.IMsgHandle

	//连接属性集合
	property map[string]interface{}

	//保护连接属性的锁
	//因为使用了map
	propertyLock sync.RWMutex
}

// NewConnection 初始化连接的方法
func NewConnection(server interfaces.IServer, conn *net.TCPConn, msgHandler interfaces.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		isClosed:     false,
		msgChan:      make(chan []byte, config.GlobalServerConfig.MaxMsgChanLen),
		MsgHandler:   msgHandler,
		property:     make(map[string]interface{}),
		propertyLock: sync.RWMutex{},
	}
	return c
}

/**
 * 启动连接，开始工作
 */
func (c *Connection) Start() {
	//赋值
	c.ctx, c.cancel = context.WithCancel(context.Background())
	logrus.Debug("Conn Start()...ConnID=", c.ConnID)
	//启动从当前连接的读数据的业务
	go c.StartReader()
	//启动从当前连接写数据的业务
	go c.StartWriter()
	//调用开发者注册的 创建连接之后 需要执行的业务Hook函数
	c.TcpServer.CallOnConnStart(c)
}

/*
StartReader 读消息的Goroutine，专门读取来自客户端的消息
*/
func (c *Connection) StartReader() {
	logrus.Debug("[Reader Goroutine is running]...")
	defer logrus.Debug("ConnID = ", c.ConnID, "[Reader is exit] ,remote addr is ", c.RemoteAddr().String())
	defer c.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			//创建拆包解包对象
			dp := c.TcpServer.Packet()
			//读取客户端的msg Head 二进制流8字节
			headData := make([]byte, dp.GetHeadLen())
			if _, err := io.ReadFull(c.Conn, headData); err != nil {
				logrus.Error("client msg head err: ", err)
				break
			}
			//拆包，得到msgID 和msgDataLen放在msg消息中
			msg, err := dp.UnPack(headData)
			if err != nil {
				logrus.Error("client unpack err: ", err)
				return
			}
			if msg == nil {
				return
			}
			//根据dataLen，再次读取Data，放在msg.Data中
			var data []byte
			if msg.GetMsgLen() > 0 {
				data = make([]byte, msg.GetMsgLen())
				if _, err := io.ReadFull(c.Conn, data); err != nil {
					logrus.Error("client read data err: ", err)
					return
				}
			}
			msg.SetMsgData(data)

			//将当前得到的conn数据封装为Request请求
			req := Request{
				conn: c,
				msg:  msg,
			}

			//从路由中找到绑定注册的conn对应的router
			//修改为根据绑定好的msgID找到对应的api处理业务
			//修改为交给worker工作池处理
			if config.GlobalServerConfig.WorkerPoolSize > 0 {
				c.MsgHandler.SendMsgToTaskQueue(&req)
			} else {
				//如果工作池未启动只能自己启动一个协程进行处理
				go c.MsgHandler.DoMsgHandle(&req)
			}
		}

	}
}

/*
StartWriter 写消息Goroutine，专门发送消息给客户端的模块
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]...")
	defer fmt.Println("ConnID = ", c.ConnID, "[Conn Writer exit] ,remote addr is ", c.RemoteAddr().String())
	//不断循环等待channel的消息
	for {
		select {
		case data := <-c.msgChan:
			//有数据写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error", err)
				return
			}
		case data, ok := <-c.msgChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					logrus.Errorf("Produce Buff Data error: %v, Conn Writer exit", err)
					return
				}
			} else {
				logrus.Debug("msgChan is Closed")
			}
			//优雅的关闭chan
		case <-c.ctx.Done():
			//代表Reader已经退出，此时Writer也要退出
			return
		}
	}
}

// SendMsg 将数据发送给channel
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	dp := c.TcpServer.Packet()
	// MsgDataLen|MsgID|MsgData 二进制数据流
	binaryMsg, err := dp.Pack(pack.NewMessage(msgId, data))
	if err != nil {
		fmt.Println("Pack msg err: ", err)
		return errors.New("pack msg error")
	}
	c.msgChan <- binaryMsg
	return nil
}

func (c *Connection) Stop() {
	c.Lock()
	defer c.Unlock()
	if c.isClosed == true {
		return
	}
	logrus.Debug("Conn stop()...ConnID=", c.ConnID)
	//调用开发者注册的 销毁连接之前 需要执行的业务Hook函数
	c.TcpServer.CallOnConnStop(c)
	//关闭socket连接
	_ = c.Conn.Close()

	//告知Writer关闭
	c.cancel()

	//将当前conn从ConnMgr中删除
	c.TcpServer.GetConnMgr().DeleteConn(c)
	//回收资源
	close(c.msgChan)
	c.isClosed = true

}
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// GetConnID 	获取连接ID
func (c *Connection) GetConnID() uint64 {
	return c.ConnID
}

func (c *Connection) SetConnID(val uint64) {
	oldId := c.ConnID
	oldConn, _ := c.TcpServer.GetConnMgr().GetConn(oldId)
	if oldConn != nil {
		//移除旧连接
		c.TcpServer.GetConnMgr().DeleteConn(oldConn)
	}
	sameConn, _ := c.TcpServer.GetConnMgr().GetConn(val)
	if sameConn != nil {
		logrus.Warnf("remove duplicated name conn[%v]", val)
		sameConn.Stop()
	}
	c.ConnID = val
	//将连接加入管理器
	c.TcpServer.GetConnMgr().AddConn(c)
}
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) Send(data []byte) error {
	return nil
}

// SetProperty 设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = value
}

// GetProperty 获取连接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	}
	return nil, errors.New("key doesn't exit")
}

// DeleteProperty 删除连接属性
func (c *Connection) DeleteProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
