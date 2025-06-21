package ratelimiter

import (
	"sync"
	"time"
)

type tokenBucketOption struct {
	timeProvider TimeProvider
}
type TokenBucketOption func(*tokenBucketOption)
type tokenBucketRateLimiter struct {
	capacity       int64
	tokens         int64
	refillRate     int64
	lastRefillTime int64
	timeProvider   TimeProvider
	mutex          sync.Mutex
}

func WithTimeProvider(provider TimeProvider) TokenBucketOption {
	return func(o *tokenBucketOption) {
		o.timeProvider = provider
	}
}

func NewTokenBucketRateLimiter(capacity int64, refillRate int64, opts ...TokenBucketOption) RateLimiter {
	opt := new(tokenBucketOption)
	for _, o := range opts {
		o(opt)
	}

	if opt.timeProvider == nil {
		opt.timeProvider = &defaultTimeProvider{}
	}

	rateLimiter := &tokenBucketRateLimiter{
		capacity:       capacity,
		tokens:         capacity,
		refillRate:     refillRate,
		timeProvider:   opt.timeProvider,
		lastRefillTime: opt.timeProvider.Now(),
		mutex:          sync.Mutex{},
	}

	return rateLimiter
}

func (t *tokenBucketRateLimiter) Allow() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.isTimeToRefill() {
		t.refill()
	}

	if t.tokens > 0 {
		t.tokens--
		return true
	}

	return false
}

func (t *tokenBucketRateLimiter) elapsedMs() int64 {
	return t.timeProvider.Now() - t.lastRefillTime
}

func (t *tokenBucketRateLimiter) timeToRefill() int64 {
	return 1000 - (t.elapsedMs())
}

func (t *tokenBucketRateLimiter) isTimeToRefill() bool {
	return t.timeToRefill() <= 0
}

func (t *tokenBucketRateLimiter) refill() {
	t.tokens = min(t.capacity, t.tokens+(t.refillRate*t.elapsedMs())/1000)
	t.lastRefillTime = t.timeProvider.Now()
}

func (t *tokenBucketRateLimiter) Wait() error {
	if t.Allow() {
		return nil
	}

	<-time.After(time.Duration(t.timeToRefill()) * time.Millisecond)
	if t.Allow() {
		return nil
	}
	return t.Wait()
}
