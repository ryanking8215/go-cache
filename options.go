package cache

import (
	"context"
	"time"
)

type Options struct {
	TTL time.Duration
	Ctx context.Context
}

func (o *Options) Apply(options ...Option) {
	o.Ctx = context.TODO()
	for _, option := range options {
		option.apply(o)
	}
}

type Option interface {
	apply(*Options)
}

type optionFunc func(*Options)

func (f optionFunc) apply(o *Options) {
	f(o)
}

// WithTTL add TTL to option
func WithTTL(ttl time.Duration) Option {
	return optionFunc(func(o *Options) {
		o.TTL = ttl
	})
}

// WithContext add Context to option
func WithContext(ctx context.Context) Option {
	return optionFunc(func(o *Options) {
		o.Ctx = ctx
	})
}
