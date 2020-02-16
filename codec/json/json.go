package json

import (
	"encoding/json"
	"errors"

	"github.com/ryanking8215/go-cache"
)

type jsonCodec struct{}

var _ cache.Codec = (*jsonCodec)(nil)

func NewCodec() *jsonCodec {
	return &jsonCodec{}
}

func (c jsonCodec) Encode(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, cache.NewCodecError(err)
	}
	return b, nil
}

func (c jsonCodec) Decode(b []byte) (interface{}, error) {
	// not support, just bypass
	return b, nil
}

func (c jsonCodec) DecodeTo(data interface{}, to interface{}) error {
	if data == nil {
		return cache.NewCodecError(errors.New("data is empty interface"))
	}
	b, ok := data.([]byte)
	if !ok {
		return cache.NewCodecError(errors.New("data is not byte slice"))
	}
	if err := json.Unmarshal(b, to); err != nil {
		return cache.NewCodecError(err)
	}
	return nil
}
