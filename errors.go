package cache

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound    = NewCacheError(errors.New("not found"))
	ErrExisted     = NewCacheError(errors.New("not existed"))
	ErrUnsupported = NewCacheError(errors.New("unsupported"))
)

type tagError struct {
	tag string
	err error
}

func newTagError(tag string, err error) *tagError {
	return &tagError{tag: tag, err: err}
}

func (e *tagError) Error() string {
	return fmt.Sprintf("%s: %v", e.tag, e.err)
}

func (e *tagError) Unwrap() error {
	return e.err
}

type CacheError struct {
	*tagError
}

func NewCacheError(err error) *CacheError {
	return &CacheError{
		tagError: newTagError("cache", err),
	}
}

type CodecError struct {
	*tagError
}

func NewCodecError(err error) *CodecError {
	return &CodecError{
		tagError: newTagError("codec", err),
	}
}
