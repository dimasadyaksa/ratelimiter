package ratelimiter

import "time"

type RateLimiter interface {
	// Allow checks if the request is allowed based on the rate limit.
	Allow() bool

	// Wait blocks until the request is allowed based on the rate limit.
	Wait() error
}

type TimeProvider interface {
	// Now returns the current time in milliseconds since epoch.
	Now() int64
}

type defaultTimeProvider struct{}

func (d *defaultTimeProvider) Now() int64 { return time.Now().UnixMilli() }
