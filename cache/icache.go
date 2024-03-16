package cache

import "time"

const (
	NeverExpire = 0
	NotSetTTL   = 1
	NotSetKey   = -2
)

type Cache interface {
	// Put 设置缓存
	Put(key string, val interface{}, timeout time.Duration) error
	// Get 取出缓存
	Get(key string) interface{}
	// Delete 删除缓存
	Delete(key string) error
	// Incr 自增
	Incr(key string) error
	// Decr 自减
	Decr(key string) error
	// IsExist 检测是否存在
	IsExist(key string) bool
	// Flush 清除缓存
	Flush() error
	// TTL TLL获取剩余的生命时间
	TTL(key string) time.Duration
}
