package local

import (
	"container/heap"
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
	LocalCacheConfig
	mu sync.Mutex
	m  map[interface{}]interface{}
	e  map[interface{}]*expireNode
	eh *expireHeap
}

var NewCache = NewLocalCache

func NewLocalCache() *localCache {
	return NewLocalCacheWithConfig(DefaultLocalCacheConfig)
}

func NewLocalCacheWithConfig(cfg LocalCacheConfig) *localCache {
	c := localCache{
		LocalCacheConfig: cfg,
		m:                make(map[interface{}]interface{}),
		e:                make(map[interface{}]*expireNode),
		eh:               &expireHeap{},
	}
	heap.Init(c.eh)

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
	for {
		n := c.eh.Pop().(*expireNode)
		if n == nil || !n.isExpired(time.Now()) || size > c.GCOnceSize {
			break
		}
		c.delNode(n)
		size++
	}
}

func (c *localCache) delNode(n *expireNode) {
	delete(c.m, n.key)
	delete(c.e, n.key)
	heap.Remove(c.eh, n.index)
}

func (c *localCache) Get(key interface{}, options ...cache.Option) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.m[key]
	if !ok {
		return nil, cache.ErrNotFound
	}
	node, ok := c.e[key]
	if ok && node.isExpired(time.Now()) {
		c.delNode(node)
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
		expireAt := time.Now().Add(o.TTL)
		n, ok := c.e[key]
		if ok { // expire node exists
			c.eh.update(n, expireAt)
		} else {
			n := &expireNode{
				key:      key,
				expireAt: expireAt,
			}
			c.e[key] = n
			c.eh.Push(n)
		}
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
		n, ok := c.e[key]
		if ok && n.isExpired(time.Now()) {
			c.delNode(n)
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
			expireAt := time.Now().Add(o.TTL)
			n, ok := c.e[k]
			if ok {
				c.eh.update(n, expireAt)
			} else {
				n := &expireNode{
					key:      k,
					expireAt: expireAt,
				}
				c.eh.Push(n)
				c.e[k] = n
			}
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
	n, ok := c.e[key]
	if ok && n.isExpired(time.Now()) {
		c.delNode(n)
		return false, nil
	}
	return true, nil
}

func (c *localCache) Delete(key interface{}, options ...cache.Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	n, ok := c.e[key]
	if ok { // expire node exists
		c.delNode(n)
	} else {
		delete(c.m, key)
	}
	return nil
}

func (c *localCache) Clear(options ...cache.Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = make(map[interface{}]interface{})
	c.e = make(map[interface{}]*expireNode)
	c.eh = &expireHeap{}
	heap.Init(c.eh)
	return nil
}

func (c *localCache) Codec() cache.Codec {
	return nil
}

type expireNode struct {
	key      interface{}
	index    int
	expireAt time.Time
}

func (n expireNode) isExpired(deadline time.Time) bool {
	if !n.expireAt.IsZero() && deadline.After(n.expireAt) {
		return true
	}
	return false
}

// An ttlHeap is a min-heap of expires
type expireHeap []*expireNode

func (h expireHeap) Len() int           { return len(h) }
func (h expireHeap) Less(i, j int) bool { return h[i].expireAt.Before(h[j].expireAt) }
func (h expireHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *expireHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	n := len(*h)
	node := x.(*expireNode)
	node.index = n
	*h = append(*h, node)
}

func (h *expireHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

func (h *expireHeap) update(n *expireNode, expireAt time.Time) {
	n.expireAt = expireAt
	heap.Fix(h, n.index)
}
