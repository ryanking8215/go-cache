package local

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ryanking8215/go-cache"
)

var _ cache.Cache = (*simpleCache)(nil)

type simpleCache struct {
	m *sync.Map
}

func NewSimpleCache() *simpleCache {
	return &simpleCache{
		m: new(sync.Map),
	}
}

func (c *simpleCache) Get(key interface{}, options ...cache.Option) (interface{}, error) {
	v, ok := c.m.Load(key)
	if !ok {
		return nil, cache.ErrNotFound
	}
	return v, nil
}

func (c *simpleCache) Set(key, value interface{}, options ...cache.Option) error {
	c.m.Store(key, value)
	return nil
}

func (c *simpleCache) MGet(keys []interface{}, options ...cache.Option) (map[interface{}]interface{}, error) {
	ret := make(map[interface{}]interface{})
	for _, key := range keys {
		if v, ok := c.m.Load(key); ok {
			ret[key] = v
		}
	}
	return ret, nil
}

func (c *simpleCache) MSet(keyValues map[interface{}]interface{}, options ...cache.Option) error {
	for k, v := range keyValues {
		c.m.Store(k, v)
	}
	return nil
}

func (c *simpleCache) Exists(key interface{}, options ...cache.Option) (bool, error) {
	_, ok := c.m.Load(key)
	return ok, nil
}

func (c *simpleCache) Delete(key interface{}, options ...cache.Option) error {
	c.m.Delete(key)
	return nil
}

func (c *simpleCache) Clear(options ...cache.Option) error {
	atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.m)), unsafe.Pointer(new(sync.Map)))
	return nil
}

func (c *simpleCache) Codec() cache.Codec {
	return nil
}
