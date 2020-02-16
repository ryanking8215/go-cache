package redis

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/ryanking8215/go-cache"
	"github.com/ryanking8215/go-cache/codec/json"
	"github.com/stretchr/testify/assert"
)

func Test_stringStoreGet(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewStringCache(rdb, json.NewCodec(), func(key string) string {
		return "test_" + key
	})

	n := 10
	for i := 0; i < n; i++ {
		if i == n-1 {
			c.Set(i, i, cache.WithTTL(time.Second))
		} else {
			c.Set(i, i)
		}
	}

	for i := 0; i < n; i++ {
		v, err := c.Get(i)
		assert.NoError(t, err)
		var val int
		err = c.Codec().DecodeTo(v, &val)
		assert.NoError(t, err)
		assert.Equal(t, i, val)
	}

	time.Sleep(2 * time.Second) // wait for expires

	for i := 0; i < n; i++ {
		v, err := c.Get(i)
		if i == n-1 {
			assert.Error(t, err)
			assert.Equal(t, cache.ErrNotFound, err)
		} else {
			assert.NoError(t, err)
			var val int
			err = c.Codec().DecodeTo(v, &val)
			assert.NoError(t, err)
			assert.Equal(t, i, val)
		}
	}
}

func Test_stringStoreSet(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewStringCache(rdb, json.NewCodec(), nil)

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

func Test_stringStoreMGet(t *testing.T) {

}

func Test_stringStoreMSet(t *testing.T) {
}

func Test_stringStoreExists(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewStringCache(rdb, json.NewCodec(), nil)

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

func Test_stringStoreDelete(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewStringCache(rdb, json.NewCodec(), nil)

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
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}
	}
}

func Test_stringStoreClear(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := NewStringCache(rdb, json.NewCodec(), nil)
	err := c.Clear()
	assert.Equal(t, cache.ErrUnsupported, err)
}
