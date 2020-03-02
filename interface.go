package cache

// Cache cache interface
type Cache interface {
	// Get Retrieves a value from cache with a specified key.
	// If key is not found, ErrNotFound will be returned.
	// Option supports cache.WithContext()
	Get(key interface{}, options ...Option) (interface{}, error)

	// Set Stores a value identified by a key into cache.
	// Option supports cache.WithContext() and cache.WithTTL()
	Set(key, value interface{}, options ...Option) error

	// MGet Retrieves multiple values from cache with the specified keys.
	// A map returned which holds the key and value pairs existed.
	// Option is same as ones of Get method
	MGet(keys []interface{}, options ...Option) (map[interface{}]interface{}, error)

	// MSet Stores multiple items in cache. Each item contains a value identified by a key.
	// Option is same as ones of Set method
	MSet(keyValues map[interface{}]interface{}, options ...Option) error

	// Exists Checks whether a specified key exists in the cache.
	// Option is same as ones of Get method
	Exists(key interface{}, options ...Option) (bool, error)

	// Delete Deletes a value with the specified key from cache.
	// Option is same as ones of Get method
	Delete(key interface{}, options ...Option) error

	// Clear Deletes all values from cache.
	// Option is same as ones of Get method
	Clear(options ...Option) error

	// Codec Retrieves codec from cache.
	Codec() Codec
}

// Encoder encoder interface
type Encoder interface {
	// Encode encode value to byte slices
	Encode(v interface{}) ([]byte, error)
}

// Decoder decoder interface
type Decoder interface {
	// Decode decode data to an instance.
	Decode(data []byte) (interface{}, error)
	// DecodeTo decode the encoded data and stores the result in the value pointed to by to
	DecodeTo(data interface{}, to interface{}) error
}

// Codec Codec is the interface that groups the Encoder and Decoder interface.
type Codec interface {
	Encoder
	Decoder
}
