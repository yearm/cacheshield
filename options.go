package cacheshield

import (
	"github.com/avast/retry-go/v4"
	"time"
)

// Option defines function type for configuration setters.
type Option func(*options)

// options contains configurable parameters for cache operations.
type options struct {
	expiration     time.Duration
	lockExpiration time.Duration
	retryOptions   []retry.Option
}

// WithExpiration sets cache expiration time.
func WithExpiration(expiration time.Duration) Option {
	return func(o *options) {
		o.expiration = expiration
	}
}

// WithLockExpiration sets lock expiration time.
func WithLockExpiration(expiration time.Duration) Option {
	return func(o *options) {
		o.lockExpiration = expiration
	}
}

// WithRetryOptions configures cache read retry behavior.
func WithRetryOptions(opts ...retry.Option) Option {
	return func(o *options) {
		o.retryOptions = opts
	}
}
