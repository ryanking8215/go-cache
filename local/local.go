package local

import (
	"sync"
	"time"

	"github.com/ryanking8215/go-cache"
)

var (
	DefaultLocalCacheConfig = LocalCacheConfig{
		GCInterval: 20 * time.Minute,
		GCOnceSize: 20,
	}
)

var _ cache.Cache = (*localCache)(nil)

type LocalCacheConfig struct {
	GCInterval time.Duration
	GCOnceSize int
}

type localCache struct {
	mu sync.Mutex
	m  map[interface{}]interface{}
	t  map[interface{}]time.Time
	LocalCacheConfig
}

var NewCache = NewLocalCache

func NewLocalCache() *localCache {
	return NewLocalCacheWithConfig(DefaultLocalCacheConfig)
}

func NewLocalCacheWithConfig(cfg LocalCacheConfig) *localCache {
	c := localCache{
		m:                make(map[interface{}]interface{}),
		t:                make(map[interface{}]time.Time),
		LocalCacheConfig: cfg,
	}

	if c.GCInterval > 0 { // Not run gc if equal 0
		go c.runGC()
	}
	return &c
}

func (c *localCache) runGC() {
	tick := time.NewTicker(c.GCInterval)
	defer func() {
		tick.Stop()
	}()

	for {
		select {
		case <-tick.C:
			c.gc()
		}
	}
}

func (c *localCache) gc() {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := 0
	for k, t := range c.t {
		if size > c.GCOnceSize {
			break
		}
		if isExpired(t, time.Now()) {
			c.del(k)
			size++
		}
	}
}

func (c *localCache) del(key interface{}) {
	delete(c.t, key)
	delete(c.m, key)
}

func (c *localCache) Get(key interface{}, options ...cache.Option) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.m[key]
	if !ok {
		return nil, cache.ErrNotFound
	}
	t, ok := c.t[key]
	if ok && isExpired(t, time.Now()) {
		c.del(key)
		return nil, cache.ErrNotFound
	}
	return v, nil
}

func (c *localCache) Set(key, value interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[key] = value
	if o.TTL > 0 {
		c.t[key] = time.Now().Add(o.TTL)
	}

	return nil
}

func (c *localCache) MGet(keys []interface{}, options ...cache.Option) (map[interface{}]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ret := make(map[interface{}]interface{})
	for _, key := range keys {
		v, ok := c.m[key]
		if !ok {
			continue
		}
		t, ok := c.t[key]
		if ok && isExpired(t, time.Now()) {
			c.del(key)
			continue
		}
		ret[key] = v
	}

	return ret, nil
}

func (c *localCache) MSet(keyValues map[interface{}]interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range keyValues {
		c.m[k] = v
		if o.TTL > 0 {
			c.t[k] = time.Now().Add(o.TTL)
		}
	}

	return nil
}

func (c *localCache) Exists(key interface{}, options ...cache.Option) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.m[key]
	if !ok {
		return false, nil
	}
	t, ok := c.t[key]
	if ok && isExpired(t, time.Now()) {
		c.del(key)
		return false, nil
	}
	return true, nil
}

func (c *localCache) Delete(key interface{}, options ...cache.Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.del(key)
	return nil
}

func (c *localCache) Clear(options ...cache.Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = make(map[interface{}]interface{})
	c.t = make(map[interface{}]time.Time)
	return nil
}

func (c *localCache) Codec() cache.Codec {
	return nil
}

func isExpired(t, deadline time.Time) bool {
	if !t.IsZero() && deadline.After(t) {
		return true
	}
	return false
}
