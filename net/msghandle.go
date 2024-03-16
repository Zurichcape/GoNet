package net

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gonet/config"
	"gonet/interfaces"
	"math/rand"
	"strconv"
)

/*
 *	消息处理模块的实现
 */

var _ interfaces.IMsgHandle = (*MsgHandle)(nil)

type MsgHandle struct {
	//存放每个msgID所对应的处理方法
	Apis map[uint32]interfaces.IRouter
	//负责Worker取消息的消息队列
	TaskQueue []chan interfaces.IRequest
	//业务工作worker池中的worker数量
	WorkerPoolSize uint
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]interfaces.IRouter),
		WorkerPoolSize: config.GlobalServerConfig.WorkerPoolSize,
		TaskQueue:      make([]chan interfaces.IRequest, config.GlobalServerConfig.WorkerPoolSize),
	}
}

// SendMsgToTaskQueue 将消息交给TaskQueue，由Worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request interfaces.IRequest) {
	//1.将消息平均分给不同的Worker
	//根据客户端建立的ConnID来进行分配，使用基本的轮询法则
	var workerID uint
	//如果此时ConnID未设置，那么就随机一个worker
	if request.GetConn().GetConnID() == 0 {
		workerID = uint(rand.Uint32()) % mh.WorkerPoolSize
	} else {
		workerID = uint(request.GetConn().GetConnID()) % mh.WorkerPoolSize
	}
	logrus.Debug("Add ConnID = ", request.GetConn().GetConnID(), "request MsgID= ",
		request.GetMsgID(), "to WorkerID= ", workerID)
	//2.将消息发送给对应的worker的TaskQueue
	mh.TaskQueue[workerID] <- request
}

// DoMsgHandle 调度/执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandle(request interfaces.IRequest) {
	//1.从Request中找到msgID
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID=" + strconv.Itoa(int(request.GetMsgID())) + "Not Found! Need Register!")
		return
	}
	//2.根据msgID调度对应的处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// AddRouter 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router interfaces.IRouter) {
	//1. 判断当前msg绑定的api的处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		//ID 已经注册了
		panic("duplicated api, msgID= " + strconv.Itoa(int(msgID)))
	}
	//2.添加msg与api的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api MsgID= ", msgID, "successful!")
}

// StartWorkerPool 启动一个worker工作池
// 开启工作池的动作只能有一次，一个GoNet框架只能有一个worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	//根据WorkerPoolSize分别开始Worker，每个Worker用一个go来承载
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//1.当前worker对应的channel消息队列，开辟空间，第0个worker就用第0个channel
		mh.TaskQueue[i] = make(chan interfaces.IRequest, config.GlobalServerConfig.MaxWorkerTaskLen)
		//2.启动当前的worker，阻塞等待消息从channel中到来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// StartOneWorker 启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan interfaces.IRequest) {
	fmt.Println("WorkerID = ", workerID, "is starting...")
	for {
		select {
		//如果有消息过来，出队列的就是一个客户端的Request，执行当前Request所绑定的业务
		case request := <-taskQueue:
			mh.DoMsgHandle(request)
		}
	}
}
