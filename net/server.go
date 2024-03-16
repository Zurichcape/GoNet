package net

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"gonet/config"
	"gonet/interfaces"
	"gonet/pack"
	"net"
)

var _ interfaces.IServer = (*Server)(nil)

// Server IServer 接口实现，定义一个Server服务类
type Server struct {
	//服务器ID
	ID uint64
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	Host string
	//服务绑定的端口
	Port int
	//消息管理模块，用来绑定MsgID和对应的处理业务api关系
	MsgHandler interfaces.IMsgHandle
	//该server的连接管理模块
	ConnMgr interfaces.IConnMgr
	//该Server创建连接之后自动调用函数-OnConnStart()
	OnConnStart func(interfaces.IConnection)
	//该Server销毁连接之前自动调用函数-OnConnStop()
	OnConnStop func(interfaces.IConnection)
	// ID生成器
	idGenerator *sonyflake.Sonyflake
	// 最大连接数
	MaxConn int
	//封/拆包方式
	packet interfaces.IDataPack
}

// NewServer 创建一个服务器句柄
func NewServer() interfaces.IServer {
	return NewServerWithParam(
		config.GlobalServerConfig.Name,
		config.GlobalServerConfig.IPVersion,
		config.GlobalServerConfig.Host,
		config.GlobalServerConfig.TCPPort,
		config.GlobalServerConfig.MaxConn,
	)
}

// NewServerWithParam 创建一个服务器句柄
func NewServerWithParam(name string, version string, host string, port int, maxConn int) interfaces.IServer {
	s := &Server{
		Name:        name,
		IPVersion:   version,
		Host:        host,
		Port:        port,
		MsgHandler:  NewMsgHandle(),
		ConnMgr:     NewConnManager(),
		packet:      pack.FactoryInstance.NewPack(interfaces.GoNetDataPack),
		MaxConn:     maxConn,
		idGenerator: NewIDGenerator(),
	}

	return s
}

// ============== 实现 IServer 里的全部接口方法 ========

func (s *Server) GetID() uint64 {
	return s.ID
}

// Start 开启网络服务
func (s *Server) Start() {
	//可以考虑做一个日志模块，将日志写到日志文件中
	logrus.Info("Server Name: %s, listener at Host: %s, Port is %d is starting ...\n", config.GlobalServerConfig.Name,
		s.Host, s.Port)
	//开启一个go去做服务端listener业务
	go func() {
		//初始化消息队列及Worker工作池
		s.MsgHandler.StartWorkerPool()

		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.Host, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}

		//2 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		//开始监听
		fmt.Println("start GoNet server  ", s.Name, " success, now listening...")
		//3 启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}

			//3.2 Server.Start() 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.GetConnLen() >= s.MaxConn {
				logrus.Debug("Too many connections MaxConn= ", s.MaxConn)
				_ = conn.Close()
				continue
			}
			//3.3 Server.Start() 处理该新连接请求的业务方法， 此时应该有 handler 和 conn是绑定的
			//server和connection集成

			dealConn := NewConnection(s, conn, s.MsgHandler)
			dealConn.SetConnID(s.GenNextID())
			//启动当前的连接业务处理
			go dealConn.Start()
		}
	}()
}

// Stop 关闭网络服务
func (s *Server) Stop() {
	//将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	logrus.Debug("[Stop] server name = ", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()
	// 阻塞，否则主Go退出，listener的go将会退出
	select {}
}
func (s *Server) AddRouter(msgID uint32, router interfaces.IRouter) {
	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("Add Router successful!")
}

func (s *Server) GetConnMgr() interfaces.IConnMgr {
	return s.ConnMgr
}

// SetOnConnStart 注册OnConnStart钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(conn interfaces.IConnection)) {
	s.OnConnStart = hookFunc
}

// SetOnConnStop 注册OnConnStop钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(conn interfaces.IConnection)) {
	s.OnConnStop = hookFunc
}

// CallOnConnStart 调用OnConnStart钩子函数的方法
func (s *Server) CallOnConnStart(conn interfaces.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("----> Call OnConnStart() ...")
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用OnConnStop钩子函数的方法
func (s *Server) CallOnConnStop(conn interfaces.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("----> Call OnConnStop() ...")
		s.OnConnStop(conn)
	}
}

// NewIDGenerator 初始化ID生成器
func NewIDGenerator() *sonyflake.Sonyflake {
	var st sonyflake.Settings
	idGen := sonyflake.NewSonyflake(st)
	if idGen == nil {
		panic("sony flake not created")
	}
	return idGen
}
func (s *Server) GenNextID() uint64 {
	id, err := s.idGenerator.NextID()
	if err != nil {
		logrus.Warnf("gonet-id-generator generates id failed: %s", err.Error())
		panic(err)
	}
	return id
}

func (s *Server) Packet() interfaces.IDataPack {
	return s.packet
}

func init() {

}
