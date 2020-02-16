package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/ryanking8215/go-cache"
	"github.com/ryanking8215/go-cache/codec/json"
	"github.com/stretchr/testify/assert"
)

func Test_hashStoreGet(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewHashCache(rdb, json.NewCodec(), "hash_test", nil)

	n := 10
	for i := 0; i < n; i++ {
		val := fmt.Sprintf("value of %d", i)
		if i == n-1 {
			c.Set(i, val, cache.WithTTL(time.Second))
		} else {
			c.Set(i, val)
		}
	}

	for i := 0; i < n; i++ {
		v, err := c.Get(i)
		assert.NoError(t, err)
		var val string
		err = c.Codec().DecodeTo(v, &val)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("value of %d", i), val)
	}

	time.Sleep(2 * time.Second) // wait for expires

	for i := 0; i < n; i++ {
		v, err := c.Get(i)
		if i == n-1 {
			assert.Error(t, err)
			assert.Equal(t, cache.ErrNotFound, err)
		} else {
			assert.NoError(t, err)
			var val string
			err = c.Codec().DecodeTo(v, &val)
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("value of %d", i), val)
		}
	}
}

func Test_hashStoreSet(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})

	c := NewHashCache(rdb, json.NewCodec(), "hash_test", nil)

	n := 10
	for i := 0; i < n; i++ {
		err := c.Set(i, i)
		assert.NoError(t, err)
	}
	for i := n; i < n*2; i++ {
		err := c.Set(i, i, cache.WithTTL(time.Minute))
		assert.NoError(t, err)
	}
}

func Test_hashStoreMGet(t *testing.T) {

}

func Test_hashStoreMSet(t *testing.T) {
}

func Test_hashStoreExists(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewHashCache(rdb, json.NewCodec(), "hash_test", nil)

	n := 10
	for i := 0; i < n; i++ {
		if i == n-1 {
			c.Set(i, i, cache.WithTTL(time.Second))
		} else {
			c.Set(i, i)
		}
	}

	for i := 0; i < n; i++ {
		ok, err := c.Exists(i)
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	time.Sleep(2 * time.Second) // wait for expires

	for i := 0; i < n; i++ {
		ok, err := c.Exists(i)
		assert.NoError(t, err)
		if i == n-1 {
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}
	}
}

func Test_hashStoreDelete(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewHashCache(rdb, json.NewCodec(), "hash_test", nil)

	n := 10
	for i := 0; i < n; i++ {
		c.Set(i, i)
	}

	for i := 0; i < n; i++ {
		if i < n/2 {
			err := c.Delete(i)
			assert.NoError(t, err)
		}
	}

	for i := 0; i < n; i++ {
		ok, err := c.Exists(i)
		assert.NoError(t, err)
		if i < n/2 {
			assert.False(t, ok, i)
		} else {
			assert.True(t, ok, i)
		}
	}
}

func Test_hashStoreClear(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewHashCache(rdb, json.NewCodec(), "hash_test", nil)
	n := 10

	for i := 0; i < n; i++ {
		c.Set(i, i)
	}

	err := c.Clear()
	assert.NoError(t, err)

	for i := 0; i < n; i++ {
		ok, _ := c.Exists(i)
		assert.False(t, ok)
	}
}
