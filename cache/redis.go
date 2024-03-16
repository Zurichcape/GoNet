package cache

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/unknwon/com"
	"gonet/db"
	"time"
)

const hsetName = "GCache"
const prefix = "c_"

const pipelinedLength = 128

// RedisCache redis缓存
type RedisCache struct {
	plr        *db.PipeLinedRedis
	hsetName   string
	prefix     string
	occupyMode bool
}

// NewRedisCache 新建缓存
func NewRedisCache() *RedisCache {
	return &RedisCache{}
}

// Initialize 初始化
// occupyMode 资源独占模式（独享整个redis）
func (c *RedisCache) Initialize(client *redis.Client, occupyMode bool) {
	c.plr = db.CreatePipeLinedRedis(client, pipelinedLength)
	c.hsetName = hsetName
	c.prefix = prefix
	c.occupyMode = occupyMode
}

// Put 设置缓存。如果expire为0，永不删除
func (c *RedisCache) Put(key string, val interface{}, expire time.Duration) error {
	key = c.prefix + key

	if err := c.plr.Set(key, com.ToStr(val), expire).Err(); err != nil {
		return err
	}

	if c.occupyMode {
		return nil
	}

	return c.plr.HSet(c.hsetName, key, "0").Err()
}

// Get 获取缓存数据
func (c *RedisCache) Get(key string) interface{} {
	val, err := c.plr.Get(c.prefix + key).Result()

	if err != nil {
		return nil
	}

	return val
}

// Delete 删除缓存数据
func (c *RedisCache) Delete(key string) error {
	key = c.prefix + key
	if err := c.plr.Del(key).Err(); err != nil {
		return err
	}

	if c.occupyMode {
		return nil
	}

	return c.plr.HDel(c.hsetName, key).Err()
}

// Incr 自增
func (c *RedisCache) Incr(key string) error {
	if !c.IsExist(key) {
		return fmt.Errorf("key '%s' not exist", key)
	}

	return c.plr.Incr(c.prefix + key).Err()
}

// Decr 自减
func (c *RedisCache) Decr(key string) error {
	if !c.IsExist(key) {
		return fmt.Errorf("key '%s' not exist", key)
	}

	return c.plr.Decr(c.prefix + key).Err()
}

// IsExist 判断是否存在
func (c *RedisCache) IsExist(key string) bool {
	if c.plr.Exists(c.prefix+key).Val() > 0 {
		return true
	}

	if !c.occupyMode {
		c.plr.HDel(c.hsetName, c.prefix+key)
	}

	return false
}

// Flush 清空所有缓存
func (c *RedisCache) Flush() error {
	if c.occupyMode {
		return c.plr.FlushAll().Err()
	}

	keys, err := c.plr.HKeys(c.hsetName).Result()
	if err != nil {
		return err
	}

	if err = c.plr.Del(keys...).Err(); err != nil {
		return err
	}

	return c.plr.Del(c.hsetName).Err()
}

// TTL 获取剩余生命时间
func (c *RedisCache) TTL(key string) time.Duration {
	ttl, err := c.plr.TTL(c.prefix + key).Result()
	if err != nil {
		return NotSetKey
	}

	return ttl
}
