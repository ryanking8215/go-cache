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

func Test_LocalCacheGet(t *testing.T) {
	c := NewLocalCache()
	n := 10
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
}

func Test_LocalCacheSet(t *testing.T) {
	c := NewLocalCache()
	n := 10
	for i := 0; i < n; i++ {
		err := c.Set(fmt.Sprintf("%d", i), i)
		assert.NoError(t, err)
	}
	for i := n; i < n*2; i++ {
		err := c.Set(fmt.Sprintf("%d", i), i, cache.WithTTL(time.Duration(i)*time.Minute))
		assert.NoError(t, err)
	}
}

func Test_LocalCacheMSet(t *testing.T) {
}

func Test_LocalCacheMGet(t *testing.T) {
}

func Test_LocalCacheExists(t *testing.T) {
	c := NewLocalCache()
	n := 10
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

func Test_LocalCacheDelete(t *testing.T) {
	n := 10
	c := NewLocalCache()
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

func Test_LocalCacheClear(t *testing.T) {
	n := 10
	c := NewLocalCache()
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
