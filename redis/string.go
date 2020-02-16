package redis

import (
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/ryanking8215/go-cache"
)

type Client = redis.UniversalClient

var _ cache.Cache = (*stringCache)(nil)

type stringCache struct {
	codec         cache.Codec
	rdb           redis.UniversalClient
	keyStringFunc func(key string) string
}

func NewStringCache(c redis.UniversalClient, codec cache.Codec, keyStringFunc func(key string) string) *stringCache {
	return &stringCache{
		rdb:           c,
		codec:         codec,
		keyStringFunc: keyStringFunc,
	}
}

func (c *stringCache) Get(key interface{}, options ...cache.Option) (interface{}, error) {
	var o cache.Options
	o.Apply(options...)

	keyStr := c.keyString(key)

	ret, err := c.rdb.DoContext(o.Ctx, "GET", keyStr).String()
	if err != nil {
		if err == redis.Nil {
			return nil, cache.ErrNotFound
		}
		return nil, cache.NewCacheError(err)
	}
	return c.codec.Decode([]byte(ret))
}

func (c *stringCache) Set(key, value interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	keyStr := c.keyString(key)
	b, err := c.codec.Encode(value)
	if err != nil {
		return err
	}

	args := make([]interface{}, 3, 5)
	args[0] = "SET"
	args[1] = keyStr
	args[2] = b
	if o.TTL > 0 {
		if usePrecise(o.TTL) {
			args = append(args, "PX", int64(o.TTL/time.Millisecond))
		} else {
			args = append(args, "EX", int64(o.TTL/time.Second))
		}
	}
	if err := c.rdb.DoContext(o.Ctx, args...).Err(); err != nil {
		return cache.NewCacheError(err)
	}
	return nil
}

func (c *stringCache) MGet(keys []interface{}, options ...cache.Option) (map[interface{}]interface{}, error) {
	var o cache.Options
	o.Apply(options...)

	args := make([]interface{}, 1, len(keys)+1)
	args[0] = "MGET"
	for _, key := range keys {
		args = append(args, c.keyString(key))
	}

	vals, err := redis.NewSliceCmd(c.rdb.DoContext(o.Ctx, args).Result()).Result()
	if err != nil {
		return nil, cache.NewCacheError(err)
	}

	ret := make(map[interface{}]interface{})
	for i, val := range vals {
		if val != nil {
			str, ok := val.(string)
			if !ok {
				continue
			}
			v, err := c.codec.Decode([]byte(str))
			if err != nil {
				continue
			}
			ret[keys[i]] = v
		}
	}

	return ret, nil
}

func (c *stringCache) MSet(keyValues map[interface{}]interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	ttlArgs := make([]interface{}, 0, 2)
	if o.TTL > 0 {
		if usePrecise(o.TTL) {
			ttlArgs = append(ttlArgs, "PX", int64(o.TTL/time.Millisecond))
		} else {
			ttlArgs = append(ttlArgs, "EX", int64(o.TTL/time.Second))
		}
	}

	pipeline := c.rdb.Pipeline()
	for k, v := range keyValues {
		keyStr := c.keyString(k)
		b, err := c.codec.Encode(v)
		if err != nil {
			return err
		}
		args := make([]interface{}, 3, 5)
		args[0] = "SET"
		args[1] = keyStr
		args[2] = b
		if len(ttlArgs) > 0 {
			args = append(args, ttlArgs...)
		}
		pipeline.Do(args)
	}

	if _, err := pipeline.ExecContext(o.Ctx); err != nil {
		return cache.NewCacheError(err)
	}

	return nil
}

func (c *stringCache) Exists(key interface{}, options ...cache.Option) (bool, error) {
	var o cache.Options
	o.Apply(options...)
	return c.rdb.DoContext(o.Ctx, "EXISTS", c.keyString(key)).Bool()
}

func (c *stringCache) Delete(key interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)
	return c.rdb.DoContext(o.Ctx, "DEL", c.keyString(key)).Err()
}

func (c *stringCache) Clear(options ...cache.Option) error {
	return cache.ErrUnsupported
}

func (c *stringCache) Codec() cache.Codec {
	return c.codec
}

func (c *stringCache) keyString(key interface{}) string {
	keyStr := toString(key, c.codec)
	if c.keyStringFunc != nil {
		keyStr = c.keyStringFunc(keyStr)
	}
	return keyStr
}
