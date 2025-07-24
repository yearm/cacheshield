package cacheshield

import (
	"context"
	"errors"
	"fmt"
	"github.com/avast/retry-go/v4"
	redisv6 "github.com/go-redis/redis"
	redisv7 "github.com/go-redis/redis/v7"
	redisv8 "github.com/go-redis/redis/v8"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/yearm/cacheshield/redis"
	goredisv6 "github.com/yearm/cacheshield/redis/goredis/v6"
	goredisv7 "github.com/yearm/cacheshield/redis/goredis/v7"
	goredisv8 "github.com/yearm/cacheshield/redis/goredis/v8"
	goredisv9 "github.com/yearm/cacheshield/redis/goredis/v9"
	"time"
)

// CallbackFunc defines the function for cache value generators.
type CallbackFunc func(ctx context.Context) (string, error)

// CacheShield provides protection against cache breakdown using distributed locks
// and automatic value regeneration. It supports multiple Redis client versions.
type CacheShield struct {
	pool redis.Pool
}

// NewV6 creates CacheShield instance for go-redis v6 client.
func NewV6(client redisv6.UniversalClient) *CacheShield {
	return &CacheShield{pool: goredisv6.NewPool(client)}
}

// NewV7 creates CacheShield instance for go-redis v7 client.
func NewV7(client redisv7.UniversalClient) *CacheShield {
	return &CacheShield{pool: goredisv7.NewPool(client)}
}

// NewV8 creates CacheShield instance for go-redis v8 client.
func NewV8(client redisv8.UniversalClient) *CacheShield {
	return &CacheShield{pool: goredisv8.NewPool(client)}
}

// NewV9 creates CacheShield instance for go-redis v9 client.
func NewV9(client redisv9.UniversalClient) *CacheShield {
	return &CacheShield{pool: goredisv9.NewPool(client)}
}

// LoadOrStore retrieves cached value or generates/store new value using callback.
func (c *CacheShield) LoadOrStore(ctx context.Context, key string, callback CallbackFunc, opts ...Option) (result string, loaded bool, err error) {
	o := options{
		expiration:     10 * time.Minute,
		lockExpiration: 10 * time.Second,
		retryOptions: []retry.Option{
			retry.Context(ctx),
			retry.Attempts(12),
			retry.Delay(200 * time.Millisecond),
			retry.MaxJitter(100 * time.Millisecond),
			retry.DelayType(retry.CombineDelay(retry.FixedDelay, retry.RandomDelay)),
			retry.LastErrorOnly(true),
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	conn := c.pool.Get(ctx)
	result, err = conn.Get(key)
	if err == nil || !errors.Is(err, redis.Nil) {
		return
	}

	mutex := NewMutex(conn, fmt.Sprintf("%s:lock", key), o.lockExpiration)
	locked, err := mutex.Lock()
	if err != nil || !locked {
		err = retry.Do(func() error {
			result, err = conn.Get(key)
			return err
		}, o.retryOptions...)
		return
	}
	defer func() { _, err = mutex.Unlock() }()

	result, err = conn.Get(key)
	if err == nil || !errors.Is(err, redis.Nil) {
		return
	}

	result, err = callback(ctx)
	if err != nil {
		return
	}
	_, err = conn.Set(key, result, o.expiration)
	if err != nil {
		return
	}
	loaded = true
	return
}

// Delete delete cached value.
func (c *CacheShield) Delete(ctx context.Context, key string) error {
	_, err := c.pool.Get(ctx).Del(key)
	if err != nil {
		return err
	}
	return nil
}
