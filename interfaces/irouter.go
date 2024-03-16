package interfaces

/*
IRouter 路由的抽象接口
路由里的数据均为Request
*/
type IRouter interface {
	// PreHandle 在处理业务之前的钩子方法Hook
	PreHandle(request IRequest)
	// Handle 处理业务的主方法hook
	Handle(request IRequest)
	// PostHandle 处理业务之后的钩子方法hook
	PostHandle(request IRequest)
}
