package interfaces

type IMsgHandle interface {

	// DoMsgHandle 调度/执行对应的Router消息处理方法
	DoMsgHandle(IRequest)

	// AddRouter 为消息添加具体的处理逻辑
	AddRouter(uint32, IRouter)

	// StartWorkerPool 启动一个worker工作池
	StartWorkerPool()

	// StartOneWorker 启动一个Worker工作流程
	StartOneWorker(int, chan IRequest)

	// SendMsgToTaskQueue 将消息交给TaskQueue，由Worker进行处理
	//需要对外暴露的方法才写在接口中
	SendMsgToTaskQueue(IRequest)
}
