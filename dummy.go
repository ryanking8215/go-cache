package cache

var _ Cache = (*dummyCache)(nil)

type dummyCache struct{}

func (c *dummyCache) Get(key interface{}, options ...Option) (interface{}, error) {
	return nil, ErrNotFound
}

func (c *dummyCache) Set(key, value interface{}, options ...Option) error {
	return nil
}

func (c *dummyCache) MGet(keys []interface{}, options ...Option) (map[interface{}]interface{}, error) {
	return nil, nil
}

func (c *dummyCache) MSet(keyValues map[interface{}]interface{}, options ...Option) error {
	return nil
}

func (c *dummyCache) Exists(key interface{}, options ...Option) (bool, error) {
	return false, nil
}

func (c *dummyCache) Delete(key interface{}, options ...Option) error {
	return nil
}

func (c *dummyCache) Clear(options ...Option) error {
	return nil
}

func (c *dummyCache) Codec() Codec {
	return nil
}
