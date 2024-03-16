package timer

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// TimeWheel 时间轮
type TimeWheel struct {
	//时间轮名称
	name string
	//刻度的时间间隔
	interval int64
	//每个时间轮上的刻度数
	scales int
	//当前时间指针的指向
	curIndex int
	//每个刻度存放的timer定时器的最大容量
	maxCap int
	//当前时间轮上的所有timer,用map来承载
	//map[int]中的int表示某个刻度数
	//map[uint32]中的uint32表示Timer的ID
	timerQueue map[int]map[uint32]*Timer
	//下一级时间轮
	nextTimeWheel *TimeWheel
	//互斥锁，保护timerQueue的操作
	sync.RWMutex
}

func NewTimeWheel(name string, interval int64, scales int, maxCap int) *TimeWheel {
	// name：时间轮的名称
	// interval：每个刻度之间的duration时间间隔
	// scales:当前时间轮的轮盘一共多少个刻度(如我们正常的时钟就是12个刻度)
	// maxCap: 每个刻度所最大保存的Timer定时器个数
	tw := &TimeWheel{
		name:     name,
		interval: interval,
		scales:   scales,
		maxCap:   maxCap,
		//初始化外层map
		timerQueue: make(map[int]map[uint32]*Timer, scales),
	}
	//初始化内层map
	for i := 0; i < maxCap; i++ {
		tw.timerQueue[i] = make(map[uint32]*Timer, maxCap)
	}
	fmt.Println("Init timeWheel name = ", tw.name, "is Done!")
	return tw
}

/*
	将一个timer定时器加入到分层时间轮中
	tID: 每个定时器timer的唯一标识
	t: 当前被加入时间轮的定时器
	forceNext: 是否强制的将定时器添加到下一层时间轮

	我们采用的算法是：
	如果当前timer的超时时间间隔 大于一个刻度，那么进行hash计算 找到对应的刻度上添加
	如果当前的timer的超时时间间隔 小于一个刻度 :
					如果没有下一轮时间轮
*/

// addTimer中操作timerQueue时没有加锁，而是放到了外部调用该方法的方法中进行
func (tw *TimeWheel) addTimer(tID uint32, t *Timer, forceNext bool) error {
	defer func() error {
		if err := recover(); err != nil {
			erst := fmt.Sprintf("addTimer function err: %s", err)
			_ = fmt.Errorf(erst)
			return errors.New(erst)
		}
		return nil
	}()
	//得到当前延迟任务的超时时间间隔(ms)
	delayInterval := t.unixts - UnixMill()

	//如果当前超时时间间隔大于一个刻度的时间间隔
	if delayInterval >= tw.interval {
		//得到需要跨越的刻度数
		dn := delayInterval / tw.interval
		//在对应的刻度上的定时器Timer集合map加入当前定时器(由于是环形，所以要求余)
		tw.timerQueue[(tw.curIndex+int(dn))%tw.scales][tID] = t
		return nil
	}
	//精度最小的时间轮
	if delayInterval < tw.interval && tw.nextTimeWheel == nil {
		if forceNext {
			// 如果设置为强制移至下一个刻度，那么将定时器移至下一个刻度
			// 这种情况，主要是时间轮自动轮转的情况
			// 因为这是底层时间轮，该定时器在转动的时候，如果没有被调度者取走的话，该定时器将不会再被发现
			// 因为时间轮刻度已经过去，如果不强制把该定时器Timer移至下时刻，就永远不会被取走并触发调用
			// 所以这里强制将timer移至下个刻度的集合中，等待调用者在下次轮转之前取走该定时器
			tw.timerQueue[(tw.curIndex+1)%tw.scales][tID] = t
		} else {
			// 如果手动添加定时器，那么直接将timer添加到对应底层时间轮的当前刻度集合中
			tw.timerQueue[tw.curIndex][tID] = t
		}
		return nil
	}
	if delayInterval < tw.interval {
		return tw.nextTimeWheel.AddTimer(tID, t)
	}
	return nil
}

// AddTimer 添加一个timer到一个时间轮中(非时间轮自转情况)
func (tw *TimeWheel) AddTimer(tID uint32, t *Timer) error {
	tw.Lock()
	defer tw.Unlock()
	return tw.addTimer(tID, t, false)
}

// RemoveTimer 删除一个定时器，根据定时器的ID
func (tw *TimeWheel) RemoveTimer(tID uint32) {
	tw.Lock()
	defer tw.Lock()
	//这里只能通过遍历每个外层来寻找tID对应的内层map
	for i := 0; i < tw.scales; i++ {
		delete(tw.timerQueue[i], tID)
	}
}

// AddTimeWheel 给一个时间轮添加下层时间轮 比如给小时时间轮添加分钟时间轮，给分钟时间轮添加秒时间轮
func (tw *TimeWheel) AddTimeWheel(nextTimeWheel *TimeWheel) {
	tw.nextTimeWheel = nextTimeWheel
	fmt.Println("Add timerWheel[", tw.name, "]'s next [", nextTimeWheel.name, "] is successful!")
}

// 启动时间轮
func (tw *TimeWheel) run() {
	for {
		time.Sleep(time.Duration(tw.interval) * time.Millisecond)
		tw.Lock()
		// 取出挂载在当前刻度的全部定时器
		curTimers := tw.timerQueue[tw.curIndex]
		// 当前定时器要重新添加 所给当前刻度再重新开辟一个map Timer容器
		tw.timerQueue[tw.curIndex] = make(map[uint32]*Timer, tw.maxCap)
		for tID, timer := range curTimers {
			// 这里属于时间轮自动转动，forceNext设置为true
			_ = tw.addTimer(tID, timer, true)
		}
		// 取出下一个刻度 挂载的全部定时器 进行重新添加 (为了安全起见,待考慮)
		nextTimers := tw.timerQueue[(tw.curIndex+1)%tw.scales]
		tw.timerQueue[(tw.curIndex+1)%tw.scales] = make(map[uint32]*Timer, tw.maxCap)
		for tID, timer := range nextTimers {
			_ = tw.addTimer(tID, timer, true)
		}

		//当前刻度指针走一个刻度
		tw.curIndex = (tw.curIndex + 1) % tw.scales

		tw.Unlock()
	}
}

func (tw *TimeWheel) Run() {
	go tw.run()
	fmt.Println("Timer wheel name = ", tw.name, "is running...")
}

// GetTimerWithin  获取定时器在一段时间间隔内的Timer
func (tw *TimeWheel) GetTimerWithin(duration time.Duration) map[uint32]*Timer {
	leaftw := tw
	// 最终触发定时器的一定是挂载最底层时间轮上的定时器
	//找到最底层的时间轮
	for leaftw.nextTimeWheel != nil {
		leaftw = leaftw.nextTimeWheel
	}
	leaftw.Lock()
	defer leaftw.Unlock()
	timerList := make(map[uint32]*Timer)
	now := UnixMill()

	// 取出当前时间轮刻度内全部Timer
	for tID, timer := range leaftw.timerQueue[leaftw.curIndex] {
		//当前定时器在duration范围内
		if timer.unixts-now < int64(duration/1e6) {
			timerList[tID] = timer
			// 定时器已经超时被取走，从当前时间轮上 摘除该定时器
			delete(leaftw.timerQueue[leaftw.curIndex], tID)
		}
	}
	return timerList
}
