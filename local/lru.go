package local

import (
	"container/list"
	"sync"
	"time"

	"github.com/ryanking8215/go-cache"
)

const (
	defaultExpired    = time.Hour
	defaultGcInterval = 10 * time.Minute
	maxGCRemovedSize  = 20
)

var DefaultLRUCacheConfig = LRUCacheConfig{
	TTL:        time.Hour,
	GCInterval: 10 * time.Minute,
	GCOnceSize: 20,
}

type LRUCacheConfig struct {
	TTL        time.Duration
	GCInterval time.Duration
	GCOnceSize int
}

var _ cache.Cache = (*lruCache)(nil)

type lruCache struct {
	Cap int
	LRUCacheConfig

	nodeList  *list.List
	nodeIndex map[interface{}]*list.Element
	mutex     sync.Mutex
}

func NewLRUCache(cap int) *lruCache {
	return NewLRUCacheWithConfig(cap, DefaultLRUCacheConfig)
}

func NewLRUCacheWithConfig(cap int, cfg LRUCacheConfig) *lruCache {
	c := &lruCache{
		nodeList:       list.New(),
		nodeIndex:      make(map[interface{}]*list.Element),
		Cap:            cap,
		LRUCacheConfig: cfg,
	}
	c.runGC()
	return c
}

func (c *lruCache) AdjustMaxCap(cap int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Cap = cap
	if c.Cap > 0 {
		for {
			if c.nodeList.Len() > c.Cap {
				e := c.nodeList.Front()
				n := e.Value.(*node)
				c.del(n.key)
			} else {
				break
			}
		}
	}
}

func (c *lruCache) runGC() {
	if c.GCInterval > 0 {
		time.AfterFunc(c.GCInterval, func() {
			c.runGC()
			c.GC()
		})
	}
}

func (c *lruCache) GC() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	removedNum := 0
	for e := c.nodeList.Front(); e != nil; {
		if removedNum > c.GCOnceSize {
			//fmt.Printf("removing %d cache nodes ..., left %d\n", removedNum, m.idList.Len())
			break
		}
		n := e.Value.(*node)
		if c.nodeIsExpired(n, time.Now()) {
			removedNum++
			next := e.Next()
			//fmt.Println("removing ...", e.Value)
			c.del(n.key)
			e = next
		}
	}
}

func (c *lruCache) Get(key interface{}, options ...cache.Option) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.get(key)
}

func (c *lruCache) Set(key interface{}, value interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.set(key, value, &o)
}

func (c *lruCache) MGet(keys []interface{}, options ...cache.Option) (map[interface{}]interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ret := make(map[interface{}]interface{})
	for _, key := range keys {
		v, err := c.get(key)
		if err != nil {
			continue
		}
		ret[key] = v
	}
	return ret, nil
}

func (c *lruCache) MSet(keyValues map[interface{}]interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for k, v := range keyValues {
		if err := c.set(k, v, &o); err != nil {
			return err
		}
	}

	return nil
}

func (c *lruCache) Exists(key interface{}, options ...cache.Option) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	el, ok := c.nodeIndex[key]
	if !ok {
		return false, nil
	}
	n := el.Value.(*node)
	if c.nodeIsExpired(n, time.Now()) {
		c.del(key)
		return false, nil
	}
	return true, nil
}

func (c *lruCache) Delete(key interface{}, options ...cache.Option) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.del(key)
}

func (c *lruCache) Clear(options ...cache.Option) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.nodeIndex = make(map[interface{}]*list.Element)
	c.nodeList = list.New()
	return nil
}

func (c *lruCache) Codec() cache.Codec {
	return nil
}

func (c *lruCache) del(key interface{}) error {
	if el, ok := c.nodeIndex[key]; ok {
		c.nodeList.Remove(el)
		delete(c.nodeIndex, key)
	}
	return nil
}

func (c *lruCache) nodeIsExpired(n *node, deadline time.Time) bool {
	ttl := c.TTL
	if n.ttl > 0 {
		ttl = n.ttl
	}
	if deadline.Sub(n.lastVisit) > ttl {
		return true
	}
	return false
}

func (c *lruCache) get(key interface{}) (interface{}, error) {
	el, ok := c.nodeIndex[key]
	if !ok {
		return nil, cache.ErrNotFound
	}

	n := el.Value.(*node)
	if c.nodeIsExpired(n, time.Now()) {
		c.del(key)
		return nil, cache.ErrNotFound
	}
	n.lastVisit = time.Now()
	c.nodeList.MoveToBack(el)

	return n.value, nil
}

func (c *lruCache) set(key, value interface{}, o *cache.Options) error {
	el, ok := c.nodeIndex[key]
	if !ok {
		el = c.nodeList.PushBack(newNode(key, value, o.TTL))
		c.nodeIndex[key] = el
	}
	n := el.Value.(*node)
	n.value = value
	n.ttl = o.TTL
	n.lastVisit = time.Now()

	if c.Cap > 0 && c.nodeList.Len() > c.Cap {
		e := c.nodeList.Front()
		n := e.Value.(*node)
		c.del(n.key)
	}

	return nil
}

type node struct {
	key       interface{}
	value     interface{}
	lastVisit time.Time
	ttl       time.Duration
}

func newNode(key, value interface{}, ttl time.Duration) *node {
	return &node{key, value, time.Now(), ttl}
}
