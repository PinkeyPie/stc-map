package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"simpleServer/internal/config"
	"simpleServer/pkg/logging"
	"strings"
	"time"
)

var _ ICacheProvider = (*RedisCacheProvider)(nil)

type RedisCacheProvider struct {
	cli    redis.UniversalClient
	cache  *cache.Cache
	prefix string
	ttl    time.Duration
}

func NewRedisCacheProvider(conf *config.Config) (ICacheProvider, error) {
	if !conf.CacheConfig.Enabled {
		return nil, fmt.Errorf("disabled cache in the config")
	}

	cli := openRedisCli(conf)
	if err := cli.Ping(context.Background()).Err(); err != nil {
		logging.DefaultLogger().Infow("failed to ping redis", "err", err)
	} else {
		logging.DefaultLogger().Info("connected to redis")
	}
	return &RedisCacheProvider{
		cli: cli,
		cache: cache.New(&cache.Options{
			Redis:        cli,
			StatsEnabled: false,
		}),
		prefix: conf.CacheConfig.Prefix,
		ttl:    conf.CacheConfig.TTL,
	}, nil
}

func (r *RedisCacheProvider) Fetch(ctx context.Context, key string, value interface{}, fetchFunc FetchFunc) error {
	if key == "" {
		return ErrInvalidKey
	}
	item := cache.Item{
		Ctx:            ctx,
		Key:            r.computeKey(key),
		Value:          value,
		TTL:            r.ttl,
		SkipLocalCache: true,
	}
	if fetchFunc != nil {
		item.Do = func(item *cache.Item) (interface{}, error) {
			return fetchFunc()
		}
	}

	return r.cache.Once(&item)
}

func (r *RedisCacheProvider) Get(ctx context.Context, key string, value interface{}) error {
	if key == "" {
		return ErrInvalidKey
	}
	if err := r.cache.Get(ctx, r.computeKey(key), value); err != nil {
		return r.wrapError(err)
	}
	return nil
}

func (r *RedisCacheProvider) Set(ctx context.Context, key string, value interface{}) error {
	if key == "" {
		return ErrInvalidKey
	}
	err := r.cache.Set(&cache.Item{
		Ctx:            ctx,
		Key:            r.computeKey(key),
		Value:          value,
		TTL:            r.ttl,
		SkipLocalCache: true,
	})
	if err != nil {
		return r.wrapError(err)
	}
	return nil
}

func (r *RedisCacheProvider) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}
	key = r.computeKey(key)
	exists := r.cache.Exists(ctx, key)

	return exists, nil
}

func (r *RedisCacheProvider) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	if err := r.cache.Delete(ctx, r.computeKey(key)); err != nil {
		return r.wrapError(err)
	}
	return nil
}

func (r *RedisCacheProvider) Close() error {
	if r.cli != nil {
		return r.cli.Close()
	}
	return nil
}

func (r *RedisCacheProvider) computeKey(key string) string {
	return r.prefix + key
}

func (r *RedisCacheProvider) wrapError(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case cache.ErrCacheMiss:
		return ErrCacheMiss
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "unkwnown compression method"):
		return ErrInvalidValue
	}
	return err
}

func openRedisCli(conf *config.Config) redis.UniversalClient {
	var (
		rconf = conf.CacheConfig.RedisConfig
	)
	if !rconf.Cluster {
		return redis.NewClient(&redis.Options{
			Addr:         rconf.Endpoints[0],
			ReadTimeout:  rconf.ReadTimeout,
			WriteTimeout: rconf.WriteTimeout,
			DialTimeout:  rconf.DialTimeout,
			PoolSize:     rconf.PoolSize,
			PoolTimeout:  rconf.PoolTimeout,
			MaxConnAge:   rconf.MaxConnAge,
			IdleTimeout:  rconf.IdleTimeout,
		})
	}
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         rconf.Endpoints,
		ReadTimeout:   rconf.ReadTimeout,
		WriteTimeout:  rconf.WriteTimeout,
		DialTimeout:   rconf.DialTimeout,
		PoolSize:      rconf.PoolSize,
		PoolTimeout:   rconf.PoolTimeout,
		MaxConnAge:    rconf.MaxConnAge,
		IdleTimeout:   rconf.IdleTimeout,
		ReadOnly:      true, // read on slave nodes.
		RouteRandomly: true, // read on masster or slave nodes.
	})
}
