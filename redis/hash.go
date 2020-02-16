package redis

import (
	"errors"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/ryanking8215/go-cache"
)

type HashCacheConfig struct {
	GCInterval time.Duration
	GCOnceSize int
}

var DefaultHashCacheConfig = HashCacheConfig{
	GCInterval: 20 * time.Minute,
	GCOnceSize: 20,
}

var _ cache.Cache = (*hashCache)(nil)

type hashCache struct {
	codec      cache.Codec
	rdb        redis.UniversalClient
	keyName    string
	timeoutKey string
	HashCacheConfig
}

func NewHashCache(rdb redis.UniversalClient, codec cache.Codec, keyName string, cfg *HashCacheConfig) *hashCache {
	c := hashCache{
		rdb:             rdb,
		codec:           codec,
		keyName:         keyName,
		timeoutKey:      keyName + ".timeout",
		HashCacheConfig: DefaultHashCacheConfig,
	}
	if cfg != nil {
		c.HashCacheConfig = *cfg
	}
	if c.GCInterval > 0 {
		go c.runGC()
	}

	return &c
}

func (c *hashCache) runGC() {
	tick := time.NewTicker(c.GCInterval)
	defer tick.Stop()

	for range tick.C {
		c.GC()
	}
}

func (c *hashCache) GC() {
	cmd := c.rdb.ZRangeByScore(c.timeoutKey, &redis.ZRangeBy{
		"-inf", timeUnixNanoToString(time.Now()), 0, 0,
	})
	fields, err := cmd.Result()
	if err != nil {
		return
	}
	_ = c.rdb.HDel(c.keyName, fields...)
}

func (c *hashCache) Get(key interface{}, options ...cache.Option) (interface{}, error) {
	var o cache.Options
	o.Apply(options...)

	field := toString(key, c.codec)
	pipeline := c.rdb.Pipeline()

	cmd := pipeline.HGet(c.keyName, field)
	scoreCmd := pipeline.ZScore(c.timeoutKey, field)
	if _, err := pipeline.ExecContext(o.Ctx); err != nil {
		if notRedisError(err) {
			return nil, err
		}
	}

	ret, err := cmd.Result()
	if err != nil {
		if err == redis.Nil {
			return nil, cache.ErrNotFound
		}
		return nil, cache.NewCacheError(err)
	}
	expire, err := scoreCmd.Result()
	if err != nil {
		if err == redis.Nil {
			expire = 0
		} else {
			return nil, cache.NewCacheError(err)
		}
	}
	if expire > 0 && int64(expire) < time.Now().UnixNano() {
		_ = c.del(&o, field)
		return nil, cache.ErrNotFound
	}

	return c.codec.Decode([]byte(ret))
}

func (c *hashCache) Set(key, value interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	field := toString(key, c.codec)
	b, err := c.codec.Encode(value)
	if err != nil {
		return err
	}

	pipeline := c.rdb.Pipeline()
	pipeline.HSet(c.keyName, field, b)
	if o.TTL > 0 {
		t := time.Now().Add(o.TTL)
		pipeline.ZAdd(c.timeoutKey, &redis.Z{Score: float64(t.UnixNano()), Member: field})
	}
	if _, err := pipeline.ExecContext(o.Ctx); err != nil {
		if notRedisError(err) {
			return err
		}
		return cache.NewCacheError(err)
	}

	return nil
}

func (c *hashCache) MGet(keys []interface{}, options ...cache.Option) (map[interface{}]interface{}, error) {
	var o cache.Options
	o.Apply(options...)

	fields := make([]string, 0, len(keys))
	for _, key := range keys {
		fields = append(fields, toString(key, c.codec))
	}

	pipeline := c.rdb.Pipeline()
	cmd := pipeline.HMGet(c.keyName, fields...)
	scoreCmds := make([]*redis.FloatCmd, 0, len(fields))
	for _, field := range fields {
		scoreCmds = append(scoreCmds, pipeline.ZScore(c.timeoutKey, field))
	}
	if _, err := pipeline.ExecContext(o.Ctx); err != nil {
		if notRedisError(err) {
			return nil, err
		}
	}

	vals, err := cmd.Result()
	if err != nil {
		return nil, cache.NewCacheError(err)
	}
	if len(vals) != len(fields) {
		return nil, cache.NewCacheError(errors.New("count not match"))
	}

	ret := make(map[interface{}]interface{})
	for i, key := range keys {
		if vals[i] == nil {
			continue
		}
		timeout, err := scoreCmds[i].Result()
		if err != nil && err != redis.Nil {
			return nil, cache.NewCacheError(err)
		}
		if timeout > 0 && int64(timeout) < time.Now().UnixNano() {
			continue
		}

		b, ok := vals[i].([]byte)
		if !ok {
			continue
		}
		v, err := c.codec.Decode(b)
		if err != nil {
			continue
		}

		ret[key] = v
	}

	return ret, nil
}

func (c *hashCache) MSet(keyValues map[interface{}]interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	fieldVals := make([]interface{}, 0, len(keyValues)*2)
	zmembers := make([]*redis.Z, 0, len(keyValues))
	var expire time.Time
	if o.TTL > 0 {
		expire = time.Now().Add(o.TTL)
	}

	pipeline := c.rdb.Pipeline()
	for k, v := range keyValues {
		b, err := c.codec.Encode(v)
		if err != nil {
			return err
		}
		field := toString(k, c.codec)
		fieldVals = append(fieldVals, field)
		fieldVals = append(fieldVals, b)

		if o.TTL > 0 {
			zmembers = append(zmembers, &redis.Z{float64(expire.UnixNano()), field})
		}
	}

	pipeline.HMSet(c.keyName, fieldVals)
	if o.TTL > 0 {
		pipeline.ZAdd(c.timeoutKey, zmembers...)
	}
	if _, err := pipeline.ExecContext(o.Ctx); err != nil {
		if notRedisError(err) {
			return err
		}
		return cache.NewCacheError(err)
	}

	return nil
}

func (c *hashCache) Exists(key interface{}, options ...cache.Option) (bool, error) {
	var o cache.Options
	o.Apply(options...)

	field := toString(key, c.codec)

	pipeline := c.rdb.Pipeline()
	cmd := pipeline.HExists(c.keyName, field)
	scoreCmd := pipeline.ZScore(c.timeoutKey, field)
	if _, err := pipeline.ExecContext(o.Ctx); err != nil {
		if notRedisError(err) {
			return false, err
		}
	}

	ret, err := cmd.Result()
	if err != nil {
		return false, cache.NewCacheError(err)
	}
	if !ret {
		return false, nil
	}

	expire, err := scoreCmd.Result()
	if err != nil && err != redis.Nil {
		return false, cache.NewCacheError(err)
	}
	if expire > 0 && int64(expire) < time.Now().UnixNano() {
		_ = c.del(&o, field)
		return false, nil
	}

	return true, nil
}

func (c *hashCache) Delete(key interface{}, options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)

	field := toString(key, c.codec)
	return c.del(&o, field)
}

func (c *hashCache) del(o *cache.Options, fields ...string) error {
	pipeline := c.rdb.Pipeline()
	pipeline.HDel(c.keyName, fields...)
	zmembers := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		zmembers = append(zmembers, field)
	}
	pipeline.ZRem(c.timeoutKey, zmembers...)
	if _, err := pipeline.ExecContext(o.Ctx); err != nil {
		if notRedisError(err) {
			return err
		}
	}
	return nil
}

func (c *hashCache) Clear(options ...cache.Option) error {
	var o cache.Options
	o.Apply(options...)
	return c.rdb.DoContext(o.Ctx, "DEL", c.keyName, c.timeoutKey).Err()
}

func (c *hashCache) Codec() cache.Codec {
	return c.codec
}
