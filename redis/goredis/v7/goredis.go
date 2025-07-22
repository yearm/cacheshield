package goredis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v7"
	credis "github.com/yearm/cacheshield/redis"
	"strings"
	"time"
)

type pool struct {
	client redis.UniversalClient
}

func (p *pool) Get(ctx context.Context) credis.Conn {
	c := p.client
	if ctx != nil {
		switch client := p.client.(type) {
		case *redis.Client:
			c = client.WithContext(ctx)
		case *redis.ClusterClient:
			c = client.WithContext(ctx)
		case *redis.Ring:
			c = client.WithContext(ctx)
		}
	}
	return &conn{client: c}
}

// NewPool returns a go-redis-based pool implementation.
func NewPool(client redis.UniversalClient) credis.Pool {
	return &pool{client: client}
}

type conn struct {
	client redis.UniversalClient
}

func (c *conn) Get(key string) (string, error) {
	value, err := c.client.Get(key).Result()
	return value, c.toError(err)
}

func (c *conn) Set(key string, value string, expiration time.Duration) (bool, error) {
	reply, err := c.client.Set(key, value, expiration).Result()
	return reply == "OK", c.toError(err)
}

func (c *conn) SetNX(key string, value string, expiration time.Duration) (bool, error) {
	b, err := c.client.SetNX(key, value, expiration).Result()
	return b, c.toError(err)
}

func (c *conn) Eval(script *credis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	keys := make([]string, script.KeyCount)
	args := keysAndArgs
	if script.KeyCount > 0 {
		for i := 0; i < script.KeyCount; i++ {
			keys[i] = keysAndArgs[i].(string)
		}
		args = keysAndArgs[script.KeyCount:]
	}

	v, err := c.client.EvalSha(script.Hash, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		v, err = c.client.Eval(script.Src, keys, args...).Result()
	}
	return v, c.toError(err)
}

func (c *conn) Del(keys ...string) (int64, error) {
	count, err := c.client.Del(keys...).Result()
	return count, c.toError(err)
}

func (c *conn) toError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, redis.Nil) {
		return credis.Nil
	}
	return err
}
