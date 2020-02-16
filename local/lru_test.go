package local

import (
	"fmt"
	"testing"
	"time"

	"github.com/ryanking8215/go-cache"
	"github.com/stretchr/testify/assert"
)

// func init() {
// 	seed := time.Now().UnixNano()
// 	rand.Seed(seed)
// }

func Test_LRUCacheGet(t *testing.T) {
	n := 10
	c := NewLRUCache(n)
	for i := 0; i < n; i++ {
		if i == 0 {
			c.Set(fmt.Sprintf("%d", i), i, cache.WithTTL(time.Second))
		} else {
			c.Set(fmt.Sprintf("%d", i), i)
		}
	}

	for i := 0; i < n; i++ {
		v, err := c.Get(fmt.Sprintf("%d", i))
		assert.NoError(t, err)
		ret := v.(int)
		assert.Equal(t, i, ret)
	}

	time.Sleep(2 * time.Second) // wait for expires

	for i := 0; i < n; i++ {
		v, err := c.Get(fmt.Sprintf("%d", i))
		if i == 0 {
			assert.Equal(t, cache.ErrNotFound, err)
		} else {
			assert.NoError(t, err)
			ret := v.(int)
			assert.Equal(t, i, ret)
		}
	}

	// evict the old
	for i := n; i < n*2; i++ {
		c.Set(fmt.Sprintf("%d", i), i)
	}
	for i := 0; i < n; i++ {
		ok, _ := c.Exists(fmt.Sprintf("%d", i))
		assert.False(t, ok)
	}
}

func Test_LRUCacheSet(t *testing.T) {
	n := 10
	c := NewLRUCache(10)
	for i := 0; i < n; i++ {
		err := c.Set(fmt.Sprintf("%d", i), i)
		assert.NoError(t, err)
	}
	for i := n; i < n*2; i++ {
		err := c.Set(fmt.Sprintf("%d", i), i, cache.WithTTL(time.Duration(i)*time.Minute))
		assert.NoError(t, err)
	}
}

func Test_LRUCacheMSet(t *testing.T) {
}

func Test_LRUCacheMGet(t *testing.T) {
}

func Test_LRUCacheExists(t *testing.T) {
	n := 10
	c := NewLRUCache(n)
	for i := 0; i < n; i++ {
		if i == 0 {
			c.Set(fmt.Sprintf("%d", i), i, cache.WithTTL(time.Second))
		} else {
			c.Set(fmt.Sprintf("%d", i), i)
		}
	}

	for i := 0; i < n; i++ {
		ok, err := c.Exists(fmt.Sprintf("%d", i))
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	time.Sleep(2 * time.Second) // wait for expires

	for i := 0; i < n; i++ {
		ok, err := c.Exists(fmt.Sprintf("%d", i))
		assert.NoError(t, err)
		if i == 0 {
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}
	}
}

func Test_LRUCacheDelete(t *testing.T) {
	n := 10
	c := NewLRUCache(n)
	for i := 0; i < n; i++ {
		c.Set(i, fmt.Sprintf("%d", i))
	}
	for i := 0; i < n; i++ {
		ok, err := c.Exists(i)
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	for i := 0; i < n/2; i++ {
		err := c.Delete(i)
		assert.NoError(t, err)
	}
	for i := 0; i < n; i++ {
		ok, err := c.Exists(i)
		assert.NoError(t, err)
		if i < n/2 {
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}
	}

	// no error when delete items not existed in cache
	for i := n; i < n*2; i++ {
		err := c.Delete(i)
		assert.NoError(t, err)
	}
}

func Test_LRUCacheClear(t *testing.T) {
	n := 10
	c := NewLRUCache(n)
	for i := 0; i < n; i++ {
		c.Set(i, fmt.Sprintf("%d", i))
	}
	for i := 0; i < n; i++ {
		ok, err := c.Exists(i)
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	assert.NoError(t, c.Clear())

	for i := 0; i < n; i++ {
		ok, err := c.Exists(i)
		assert.NoError(t, err)
		assert.False(t, ok)
	}
}
