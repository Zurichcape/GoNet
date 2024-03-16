package timer

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	// MaxChanBuff 默认缓冲触发函数队列大小
	MaxChanBuff = 2048
	// MaxDelayTime 默认最大误差时间
	MaxDelayTime = 100
)

// TimerScheduler 计时器调度器
type TimerScheduler struct {
	//当前调度器的最高级时间轮
	tw *TimeWheel
	//定时器编号累加器
	IDGen uint32
	//已经触发定时器的延迟方法的channel
	triggerChan chan *DelayFunc
	//互斥锁
	sync.RWMutex
	//所有注册的timerID集合
	IDs []uint32
}

// NewTimerScheduler 返回一个定时器调度器 ，主要创建分层定时器，并做关联，并依次启动
func NewTimerScheduler() *TimerScheduler {
	// 创建秒级时间轮
	secondTimeWheel := NewTimeWheel(SecondName, SecondInterval, SecondScales, TimersMaxCap)
	// 创建分钟级时间轮
	minuteTimeWheel := NewTimeWheel(MinuteName, MinuteInterval, MinuteScales, TimersMaxCap)
	// 创建小时级时间轮
	hourTimeWheel := NewTimeWheel(HourName, HourInterval, HourScales, TimersMaxCap)

	//分层时间轮做关联
	hourTimeWheel.nextTimeWheel = minuteTimeWheel
	minuteTimeWheel.nextTimeWheel = secondTimeWheel

	//时间轮运行
	secondTimeWheel.Run()
	minuteTimeWheel.Run()
	hourTimeWheel.Run()

	return &TimerScheduler{
		tw:          hourTimeWheel,
		triggerChan: make(chan *DelayFunc, MaxChanBuff),
		IDs:         make([]uint32, 0),
	}
}

// CreateTimerAt 创建一个定点Timer 并将Timer添加到分层时间轮中， 返回Timer的tID
func (ts *TimerScheduler) CreateTimerAt(df *DelayFunc, unixNano int64) (uint32, error) {
	ts.Lock()
	defer ts.Unlock()
	ts.IDGen++
	ts.IDs = append(ts.IDs, ts.IDGen)
	return ts.IDGen, ts.tw.AddTimer(ts.IDGen, NewTimerAt(df, unixNano))
}

// CreateTimerAfter 创建一个延迟Timer 并将Timer添加到分层时间轮中， 返回Timer的tID
func (ts *TimerScheduler) CreateTimerAfter(df *DelayFunc, duration time.Duration) (uint32, error) {
	ts.Lock()
	defer ts.Unlock()
	ts.IDGen++
	ts.IDs = append(ts.IDs, ts.IDGen)
	return ts.IDGen, ts.tw.AddTimer(ts.IDGen, NewTimerAfter(df, duration))
}

// CancelTimer 删除timer
func (ts *TimerScheduler) CancelTimer(tID uint32) {
	ts.Lock()
	defer ts.Unlock()
	//删除timerID
	var index = -1
	for i := 0; i < len(ts.IDs); i++ {
		if ts.IDs[i] == tID {
			index = i
		}
	}
	//直接从IDs里删除，而不需要从底层时间轮中删除，通过tID来进行调度
	if index > -1 {
		ts.IDs = append(ts.IDs[:index], ts.IDs[index+1:]...)
	}
}

// GetTriggerChan 获取计时结束的延迟执行函数通道
func (ts *TimerScheduler) GetTriggerChan() chan *DelayFunc {
	return ts.triggerChan
}

// HasTimer 是否有时间轮
func (ts *TimerScheduler) HasTimer(tID uint32) bool {
	for i := 0; i < len(ts.IDs); i++ {
		if ts.IDs[i] == tID {
			return true
		}
	}
	return false
}

// Start 非阻塞式启动
func (ts *TimerScheduler) Start() {
	go func() {
		for {
			now := UnixMill()
			// 获取最近MaxTimeDelay 毫秒的超时定时器集合
			timerList := ts.tw.GetTimerWithin(MaxDelayTime * time.Millisecond)
			for tID, timer := range timerList {
				if math.Abs(float64(now-timer.unixts)) > MaxDelayTime {
					// 已经超时的定时器，报警
					_ = fmt.Errorf("want call at: %v; real call at: %v; delay %v", timer.unixts, now, now-timer.unixts)
				}
				if ts.HasTimer(tID) {
					// 将超时触发函数写入管道
					ts.triggerChan <- timer.delayFunc
				}
			}
			//每隔50微秒进行读取1次过期的timer
			time.Sleep(MaxDelayTime / 2 * time.Millisecond)
		}
	}()
}

// NewAutoExecTimerScheduler 时间轮定时器 自动调度
func NewAutoExecTimerScheduler() *TimerScheduler {
	// 创建一个调度器
	autoExecScheduler := NewTimerScheduler()
	// 启动调度器
	autoExecScheduler.Start()

	// 永久从调度器中获取超时触发的函数 并执行
	go func() {
		delayFuncChan := autoExecScheduler.GetTriggerChan()
		for df := range delayFuncChan {
			go df.Call()
		}
	}()
	return autoExecScheduler
}
