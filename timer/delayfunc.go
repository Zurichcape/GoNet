package timer

import (
	"fmt"
	"reflect"
)

/*
	    定义一个延迟调用函数
		延迟调用函数就是 时间定时器超时的时候，触发的事先注册好的
		回调函数
*/
type DelayFunc struct {
	f    func(...interface{}) //延迟调用函数原型
	args []interface{}        //延迟调用函数的参数
}

// NewDelayFunc 创建一个延迟调用函数
func NewDelayFunc(f func(v ...interface{}), args []interface{}) *DelayFunc {
	return &DelayFunc{
		f:    f,
		args: args,
	}
}

// String 打印当前延迟函数的信息，用于日志记录
func (df *DelayFunc) String() string {
	return fmt.Sprintf("{DelayFunc:%s, args:%v}", reflect.TypeOf(df.f).Name(), df.args)
}

func (df *DelayFunc) Call() {
	defer func() {
		if err := recover(); err != nil {
			_ = fmt.Errorf("%v Call error: %v", df.String(), err)
		}
	}()
	//调用定时器超时函数
	df.f(df.args...)
}
