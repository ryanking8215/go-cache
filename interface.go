package cache

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

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

type Decoder interface {
	Decode(data []byte) (interface{}, error)
	DecodeTo(data interface{}, to interface{}) error
}

type Codec interface {
	Encoder
	Decoder
}
