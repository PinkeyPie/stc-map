package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"simpleServer/internal/config"
)

var (
	ErrCacheMiss    = errors.New("key is missing")
	ErrInvalidKey   = errors.New("key is invalid")
	ErrInvalidValue = errors.New("value type is invalid")
)

type FetchFunc func() (interface{}, error)

type ICacheProvider interface {
	io.Closer
	Fetch(ctx context.Context, key string, value interface{}, fetchFunc FetchFunc) error
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}) error
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
}

func NewCacheProvider(conf *config.Config) (ICacheProvider, error) {
	fmt.Println("===================================== ICacheProvider create =========================================================")
	if !conf.CacheConfig.Enabled {
		return nil, nil
	}
	switch conf.CacheConfig.Type {
	case "redis":
		return NewRedisCacheProvider(conf)
	default:
		return nil, fmt.Errorf("unknown cache type: %s", conf.CacheConfig.Type)
	}
}

type contextKey = string

const skipCacheKey = contextKey("skipCacheKey")

func IsCacheSkip(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if skip, exists := ctx.Value(skipCacheKey).(bool); exists {
		return skip
	}
	return false
}

func WithCacheSkip(ctx context.Context, skipCache bool) context.Context {
	return context.WithValue(ctx, skipCacheKey, skipCache)
}
