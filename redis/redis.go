package redis

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"time"
)

// Nil reply returned by Redis when key does not exist.
var Nil = errors.New("redis: nil")

// Pool maintains a pool of Redis connections.
type Pool interface {
	Get(ctx context.Context) Conn
}

// Conn is a single Redis connection.
type Conn interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) (bool, error)
	SetNX(key string, value string, expiration time.Duration) (bool, error)
	Eval(script *Script, keysAndArgs ...interface{}) (interface{}, error)
	Del(keys ...string) (int64, error)
}

// Script encapsulates the source, hash and key count for a Lua script.
type Script struct {
	KeyCount int
	Src      string
	Hash     string
}

// NewScript returns a new script object. If keyCount is greater than or equal
// to zero, then the count is automatically inserted in the EVAL command
// argument list. If keyCount is less than zero, then the application supplies
// the count as the first value in the keysAndArgs argument to the Do, Send and
// SendHash methods.
func NewScript(keyCount int, src string) *Script {
	h := sha1.New()
	_, _ = io.WriteString(h, src)
	return &Script{
		KeyCount: keyCount,
		Src:      src,
		Hash:     hex.EncodeToString(h.Sum(nil)),
	}
}
