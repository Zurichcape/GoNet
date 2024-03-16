package db

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gonet/utils"
	"sync"
	"time"
)

type redisOpName int

const (
	setOp redisOpName = iota
	getOp
	delOp
	setnxOp
	existsOp
	incrOp
	decrOp
	flushAllOp
	ttlOp
	expireOp
	hsetOp
	hgetOp
	hdelOp
	hkeysOp
	saddOp
	smembersOp
	scardOp
	lpushOp
	ltrimOp
	llenOp
	lrangeOp
)

type PipeLinedRedis struct {
	client         *redis.Client
	opChan         chan *redisOp
	pipeLineLength int
	exitFlag       chan bool
	sync.Mutex
}

type redisOp struct {
	op       redisOpName
	key      string
	keys     []string
	members  []interface{}
	hKey     string
	val      interface{}
	expire   time.Duration
	cmd      redis.Cmder
	isFinish chan bool
	start    int64
	end      int64
}

func (op *redisOp) finish(cmder redis.Cmder) {
	op.cmd = cmder
	op.isFinish <- true
}

func (op *redisOp) exec(ctx context.Context, pipeline redis.Pipeliner) {
	switch op.op {
	case setOp:
		pipeline.Set(ctx, op.key, op.val, op.expire)
	case getOp:
		op.cmd = pipeline.Get(ctx, op.key)
	case delOp:
		op.cmd = pipeline.Del(ctx, op.keys...)
	case setnxOp:
		op.cmd = pipeline.SetNX(ctx, op.key, op.val, 0)
	case existsOp:
		op.cmd = pipeline.Exists(ctx, op.keys...)
	case incrOp:
		op.cmd = pipeline.Incr(ctx, op.key)
	case decrOp:
		op.cmd = pipeline.Decr(ctx, op.key)
	case flushAllOp:
		op.cmd = pipeline.FlushAll(ctx)
	case ttlOp:
		op.cmd = pipeline.TTL(ctx, op.key)
	case expireOp:
		op.cmd = pipeline.Expire(ctx, op.key, op.expire)
	case hsetOp:
		op.cmd = pipeline.HSet(ctx, op.key, op.hKey, op.val)
	case hgetOp:
		op.cmd = pipeline.HGet(ctx, op.key, op.hKey)
	case hdelOp:
		op.cmd = pipeline.HDel(ctx, op.key, op.hKey)
	case hkeysOp:
		op.cmd = pipeline.HKeys(ctx, op.key)
	case saddOp:
		op.cmd = pipeline.SAdd(ctx, op.key, op.members...)
	case scardOp:
		op.cmd = pipeline.SCard(ctx, op.key)
	case smembersOp:
		op.cmd = pipeline.SMembers(ctx, op.key)
	case lpushOp:
		op.cmd = pipeline.LPush(ctx, op.key, op.members...)
	case ltrimOp:
		op.cmd = pipeline.LTrim(ctx, op.key, op.start, op.end)
	case llenOp:
		op.cmd = pipeline.LLen(ctx, op.key)
	case lrangeOp:
		op.cmd = pipeline.LRange(ctx, op.key, op.start, op.end)
	}
}

func CreatePipeLinedRedis(client *redis.Client, pipeLineLength int) *PipeLinedRedis {
	plr := &PipeLinedRedis{
		client:         client,
		opChan:         make(chan *redisOp, pipeLineLength*8),
		pipeLineLength: pipeLineLength,
		exitFlag:       make(chan bool, 1),
	}

	go plr.runLoop()

	return plr
}

func (plr *PipeLinedRedis) Exit() {
	plr.exitFlag <- true
}

func (plr *PipeLinedRedis) runLoop() {
	for {
		// 等待通道有值
		select {
		case op := <-plr.opChan:
			opList := make([]*redisOp, plr.pipeLineLength)
			opList[0] = op
			// 有值的时候，循环取出，直到取空或者取够pipeLineLength的数量
			isEmpty := false
			i := 1
			for ; i < plr.pipeLineLength && !isEmpty; i++ {
				select {
				case op = <-plr.opChan:
					opList[i] = op
				default:
					isEmpty = true
				}
			}
			opList = opList[:i-1]
			// todo 优化效果，对相同的操作去重
			// 加锁执行操作
			plr.Lock()
			ctx, cancel := utils.MakeCtx(time.Minute)
			pipeLine := plr.client.Pipeline()
			for _, op := range opList {
				op.exec(ctx, pipeLine)
			}
			cmdList, _ := pipeLine.Exec(ctx)
			for j, op := range opList {
				op.finish(cmdList[j])
			}
			cancel()
			plr.Unlock()
		case exit := <-plr.exitFlag:
			if exit {
				return
			}
		}
	}
}

func (plr *PipeLinedRedis) Set(key string, val interface{}, expire time.Duration) *redis.StatusCmd {
	op := &redisOp{
		op:       setOp,
		key:      key,
		val:      val,
		expire:   expire,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StatusCmd)
}

func (plr *PipeLinedRedis) Get(key string) *redis.StringCmd {
	op := &redisOp{
		op:       getOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StringCmd)
}

func (plr *PipeLinedRedis) Del(keys ...string) *redis.IntCmd {
	op := &redisOp{
		op:       delOp,
		keys:     keys,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) Setnx(key string, val interface{}) *redis.BoolCmd {
	op := &redisOp{
		op:       setnxOp,
		key:      key,
		val:      val,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.BoolCmd)
}

func (plr *PipeLinedRedis) Incr(key string) *redis.IntCmd {
	op := &redisOp{
		op:       incrOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) Decr(key string) *redis.IntCmd {
	op := &redisOp{
		op:       decrOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) Exists(keys ...string) *redis.IntCmd {
	op := &redisOp{
		op:       existsOp,
		keys:     keys,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) FlushAll() *redis.StatusCmd {
	op := &redisOp{
		op:       flushAllOp,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StatusCmd)
}

func (plr *PipeLinedRedis) TTL(key string) *redis.DurationCmd {
	op := &redisOp{
		op:       ttlOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.DurationCmd)
}

func (plr *PipeLinedRedis) Expire(key string, expire time.Duration) *redis.BoolCmd {
	op := &redisOp{
		op:       expireOp,
		key:      key,
		expire:   expire,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.BoolCmd)
}

func (plr *PipeLinedRedis) HSet(key string, hKey string, val interface{}) *redis.IntCmd {
	op := &redisOp{
		op:       hsetOp,
		key:      key,
		hKey:     hKey,
		val:      val,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) HGet(key string, hKey string) *redis.StringCmd {
	op := &redisOp{
		op:       hgetOp,
		key:      key,
		hKey:     hKey,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StringCmd)
}

func (plr *PipeLinedRedis) HDel(key string, hKey string) *redis.IntCmd {
	op := &redisOp{
		op:       hdelOp,
		key:      key,
		hKey:     hKey,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) HKeys(key string) *redis.StringSliceCmd {
	op := &redisOp{
		op:       hkeysOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StringSliceCmd)
}

func (plr *PipeLinedRedis) SAdd(key string, members []interface{}) *redis.IntCmd {
	op := &redisOp{
		op:       saddOp,
		key:      key,
		members:  members,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) SMembers(key string) *redis.StringSliceCmd {
	op := &redisOp{
		op:       smembersOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StringSliceCmd)
}

func (plr *PipeLinedRedis) SCard(key string) *redis.IntCmd {
	op := &redisOp{
		op:       scardOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) LPush(key string, members []interface{}) *redis.IntCmd {
	op := &redisOp{
		op:       lpushOp,
		key:      key,
		isFinish: make(chan bool, 1),
		members:  members,
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) LTrim(key string, start, end int64) *redis.StatusCmd {
	op := &redisOp{
		op:       ltrimOp,
		key:      key,
		isFinish: make(chan bool, 1),
		start:    start,
		end:      end,
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StatusCmd)
}

func (plr *PipeLinedRedis) LLen(key string) *redis.IntCmd {
	op := &redisOp{
		op:       llenOp,
		key:      key,
		isFinish: make(chan bool, 1),
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.IntCmd)
}

func (plr *PipeLinedRedis) LRange(key string, start, end int64) *redis.StringSliceCmd {
	op := &redisOp{
		op:       lrangeOp,
		key:      key,
		isFinish: make(chan bool, 1),
		start:    start,
		end:      end,
	}
	plr.opChan <- op
	<-op.isFinish

	return op.cmd.(*redis.StringSliceCmd)
}
