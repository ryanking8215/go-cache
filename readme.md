# Go Cache
Cache written by [Golang](https://golang.org). 

Features:
## multi implements
* local - store in local map with ttl support
* lru - local cache with lru evicted policy.
* redis string - store in redis string type.
* redis hash - store in redis hash type.
* dummy - dummy cache for placeholder.

## multi codec
* json - json encode/decode

## functional options pattern. 
provides options like TTL, context support and so on.

# Usage
## interface
Interface "Cache" defines the common cache methods.

```golang
type Cache interface {
	Get(key interface{}, options ...Option) (interface{}, error)
	Set(key, value interface{}, options ...Option) error
	MGet(keys []interface{}, options ...Option) (map[interface{}]interface{}, error)
	MSet(keyValues map[interface{}]interface{}, options ...Option) error
	Exists(key interface{}, options ...Option) (bool, error)
	Delete(key interface{}, options ...Option) error
	Clear(options ...Option) error
	Codec() Codec
}
```

## options
```golang
// no options
cache.Set("key", "value")
// ttl 
cache.Set("key", "value", cache.WithTTL(time.Minute))
// context, useful for redis cache
cache.Set("key", "value", cache.WithContext(context.WithTimeout(ctx, 2*time.Second)))
```

Feel free to combine them:
```golang
cache.Set("key", "value", cache.WithTTL(time.Minute), cache.WithContext(context.WithTimeout(ctx, 2*time.Second)))
```

## local cache
```golang
import (
    "time"

    "github.com/ryanking8215/go-cache"
    "github.com/ryanking8215/go-cache/local"
)

func main() {
    c := local.NewCache()
    n := 10
    for i:=0; i<n; i++ {
        c.Set(i, fmt.Sprintf("%d", i))
    }
    c.Set("with ttl", "5 minute", cache.WithTTL(5*time.Minute))
}
```

## local lru cache
```golang
import (
    "time"

    "github.com/ryanking8215/go-cache"
    "github.com/ryanking8215/go-cache/local"
)

func main() {
    n := 10
    c := local.NewLRUCache(10)
    for i:=0; i<n; i++ {
        c.Set(i, fmt.Sprintf("%d", i))
    }
    c.Set("with ttl", "5 minute", cache.WithTTL(5*time.Minute))
}
```

LRU cache has its default TTL for all keys when initialized, and also supports custom ttl for individual key.

## redis string cache

```golang
import (
    "time"

	"github.com/go-redis/redis/v7"
	"github.com/ryanking8215/go-cache"
	"github.com/ryanking8215/go-cache/codec/json"
    rediscache "github.com/ryanking8215/go-cache/redis"
)

func main() {
    rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := rediscache.NewStringCache(rdb, json.NewCodec(), func(key string) string{
        return "test_"+key // feel free to generate your own key layout
    })
    n := 10
    c := rediscache.NewStringCache(10)
    for i:=0; i<n; i++ {
        c.Set(i, fmt.Sprintf("%d", i))
    }
    c.Set("with ttl", "5 minute", cache.WithTTL(5*time.Minute))

    // get value with codec
    v, err := c.Get(0)
    if err!=nil {
        return
    }
    var val string
    if err:=c.Codec().DecodeTo(v, &val); err!=nil {
        return
    }
}
```

    Be aware that redis string cache doesn't support Clear() method!


## redis hash cache
```golang
import (
    "time"

	"github.com/go-redis/redis/v7"
	"github.com/ryanking8215/go-cache"
	"github.com/ryanking8215/go-cache/codec/json"
    rediscache "github.com/ryanking8215/go-cache/redis"
)

func main() {
    rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	c := rediscache.NewHashCache(rdb, json.NewCodec(), "hash_test")
    n := 10
    for i:=0; i<n; i++ {
        c.Set(i, fmt.Sprintf("%d", i))
    }
    c.Set("with ttl", "5 minute", cache.WithTTL(5*time.Minute))
}
```

## dummy cache
```golang
import (
    "time"

    "github.com/ryanking8215/go-cache"
)

func main() {
    c := cache.NewDummyCache()
    n := 10
    for i:=0; i<n; i++ {
        c.Set(i, fmt.Sprintf("%d", i))
    }
}
```

# TODO
* Add simple cache implemented by sync.Map for extreme performence usage.
* Local cache should have a better GC(release expired keys) implement.
* Redis hash cache has concurrent problem. (implemented by pipeline now, lua script may work I think, or any other good advice).
* Gob codec support.
* More tests.