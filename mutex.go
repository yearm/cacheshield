package cacheshield

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/yearm/cacheshield/redis"
	"time"
)

// deleteScript lua script for atomic lock release with value verification.
var deleteScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

// Mutex implements a distributed mutex using Redis.
type Mutex struct {
	conn       redis.Conn
	name       string
	value      string
	expiration time.Duration
}

// NewMutex creates a new distributed mutex instance.
func NewMutex(conn redis.Conn, name string, expiration time.Duration) *Mutex {
	return &Mutex{conn: conn, name: name, expiration: expiration}
}

// Lock attempts to acquire the distributed mutex.
func (m *Mutex) Lock() (bool, error) {
	value, err := m.genValue()
	if err != nil {
		return false, err
	}
	b, err := m.conn.SetNX(m.name, value, m.expiration)
	if err != nil {
		return false, err
	}
	m.value = value
	return b, nil
}

// Unlock releases the distributed mutex atomically.
func (m *Mutex) Unlock() (bool, error) {
	status, err := m.conn.Eval(deleteScript, m.name, m.value)
	if err != nil {
		return false, err
	}
	return status != int64(0), nil
}

// genValue creates a unique cryptographic token for lock ownership.
func (m *Mutex) genValue() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
