package timer

import "time"

const (
	// HourName 小时
	HourName = "HOUR"
	// HourInterval 小时时间间隔精度为ms
	HourInterval = 60 * 60 * 1e3
	// HourScales 12小时制
	HourScales = 12

	// MinuteName 分钟
	MinuteName = "MINUTE"
	// MinuteInterval 每分钟时间间隔
	MinuteInterval = 60 * 1e3
	// MinuteScales 60分钟
	MinuteScales = 60

	// SecondName 秒
	SecondName = "SECOND"
	// SecondInterval 1秒的时间间隔
	SecondInterval = 1e3
	// SecondScales 60秒
	SecondScales = 60
	// TimersMaxCap 每个时间轮刻度挂载定时器的最大个数
	TimersMaxCap = 2048
)

/*
   注意：
    有关时间的几个换算
   	time.Second(秒) = time.Millisecond * 1e3
	time.Millisecond(毫秒) = time.Microsecond * 1e3
	time.Microsecond(微秒) = time.Nanosecond * 1e3

	time.Now().UnixNano() ==> time.Nanosecond (纳秒)
*/

type Timer struct {
	//延迟调用函数
	delayFunc *DelayFunc
	//调用时间(unix时间，单位ms)
	unixts int64
}

// UnixMill 返回从1970-01-01到此时经历的毫秒数
func UnixMill() int64 {
	return time.Now().UnixNano() / 1e6
}

// NewTimerAt   创建一个定时器,在指定的时间触发
// 定时器方法 df: DelayFunc类型的延迟调用函数类型；
// unixNano: unix计算机从1970-1-1至今经历的纳秒数
func NewTimerAt(df *DelayFunc, unixNano int64) *Timer {
	return &Timer{
		delayFunc: df,
		unixts:    unixNano,
	}
}

// NewTimerAfter 创建一个定时器，在当前时间延迟duration之后触发 定时器方法
func NewTimerAfter(df *DelayFunc, duration time.Duration) *Timer {
	return &Timer{
		delayFunc: df,
		unixts:    time.Now().UnixNano() + int64(duration),
	}
}

// Run 启动定时器，用一个go承载
func (t *Timer) Run() {
	go func() {
		now := UnixMill()
		//设置的定时器是否在当前时间之后
		if t.unixts > now {
			//睡眠直到超时
			time.Sleep(time.Duration(t.unixts-now) * time.Millisecond)
		}
		//到时间了调用注册的延迟方法
		t.delayFunc.Call()
	}()
}
