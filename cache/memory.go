package cache

import (
	"errors"
	"sync"
	"time"
)

// MemoryItem 内存缓存对象
type MemoryItem struct {
	val     interface{}
	created time.Time
	expire  time.Duration
}

// hasExpired 判断缓存对象是否过期
func (item *MemoryItem) hasExpired() bool {
	return item.expire > 0 && time.Since(item.created) > item.expire
}

// MemoryCache 内存缓存
type MemoryCache struct {
	lock  sync.RWMutex
	items map[string]*MemoryItem
	//自定义GC的时间间隔
	interval int
}

// NewMemoryCache 新建一个内存缓存
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{items: make(map[string]*MemoryItem)}
}

// Initialize 初始化
func (c *MemoryCache) Initialize(interval int) error {
	c.lock.Lock()
	c.interval = interval
	c.lock.Unlock()

	go c.GC()
	return nil
}

// Put 设置一个缓存和过期时间
func (c *MemoryCache) Put(key string, val interface{}, expire time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items[key] = &MemoryItem{
		val:     val,
		created: time.Now(),
		expire:  expire,
	}
	return nil
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil
	}
	//惰性删除
	if item.hasExpired() {
		go c.Delete(key)
		return nil
	}
	return item.val
}

// Delete 删除一个缓存
func (c *MemoryCache) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.items, key)
	return nil
}

// Incr 自增
func (c *MemoryCache) Incr(key string) (err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return errors.New("key not exist")
	}

	item.val, err = Incr(item.val)

	return err
}

// Decr 自减
func (c *MemoryCache) Decr(key string) (err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return errors.New("key not exist")
	}

	item.val, err = Decr(item.val)

	return err
}

// IsExist 检查是否存在
func (c *MemoryCache) IsExist(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, ok := c.items[key]

	return ok
}

// Flush 清空所有缓存
func (c *MemoryCache) Flush() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items = make(map[string]*MemoryItem)

	return nil
}

// TTL 获取剩余生命时间
func (c *MemoryCache) TTL(key string) time.Duration {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return NotSetKey
	}

	if item.hasExpired() {
		go c.Delete(key)
		return NotSetKey
	}

	if item.expire == NeverExpire {
		return NotSetTTL
	}
	return time.Since(item.created)
}

// checkRawExpiration 检测是否过期
func (c *MemoryCache) checkRawExpiration(key string) {
	item, ok := c.items[key]
	if !ok {
		return
	}

	if item.hasExpired() {
		delete(c.items, key)
	}
}

// checkExpiration 检测是否过期
func (c *MemoryCache) checkExpiration(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.checkRawExpiration(key)
}

func (c *MemoryCache) GC() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.interval < 1 {
		return
	}

	if c.items != nil {
		for key := range c.items {
			c.checkRawExpiration(key)
		}
	}

	time.AfterFunc(time.Duration(c.interval)*time.Second, func() { c.GC() })
}
