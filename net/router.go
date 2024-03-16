package net

import "gonet/interfaces"

var _ interfaces.IRouter = (*BaseRouter)(nil)

/*
BaseRouter 先嵌入BaseRouter基类，然后根据需求对这个基类方法进行重写
*代理模式
*/
type BaseRouter struct {
}

// 这里之所以所有的BaseRouter方法均为空，是因为有的Router不需要所有的方法

// PreHandle 在处理业务之前的钩子方法Hook
func (b *BaseRouter) PreHandle(request interfaces.IRequest) {}

// Handle 处理业务的主方法hook
func (b *BaseRouter) Handle(request interfaces.IRequest) {}

// PostHandle 处理业务之后的钩子方法hook
func (b *BaseRouter) PostHandle(request interfaces.IRequest) {}
