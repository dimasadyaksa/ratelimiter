package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/dimasadyaksa/ratelimiter"
)

type mockTimeProvider struct {
	currentTime int64
}

func (m *mockTimeProvider) Now() int64 {
	return m.currentTime
}

func (m *mockTimeProvider) SetTime(t int64) {
	m.currentTime = t
}

func TestTokenBucketRateLimiter(t *testing.T) {
	timeProvider := &mockTimeProvider{}

	t.Run("Must allow requests up to capacity", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))
		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}
	})

	t.Run("Must not allow requests exceeding capacity", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))
		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}
		if limiter.Allow() {
			t.Error("Expected request to be denied after capacity exceeded")
		}
	})

	t.Run("Must not refill token before 1s", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))
		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}
		timeProvider.SetTime(999)
		if limiter.Allow() {
			t.Error("Expected request to be denied before 1 second has passed")
		}
	})

	t.Run("Must refill tokens over time", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))

		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}

		timeProvider.SetTime(1000) // Simulate time passing by 1 second
		if !limiter.Allow() {
			t.Error("Expected request to be allowed after refill")
		}
	})

	t.Run("Must not refill tokens more than capacity", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))

		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}

		timeProvider.SetTime(1000) // Simulate time passing by 1 second
		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed after refill", i)
			}
		}

		if limiter.Allow() {
			t.Error("Expected request to be denied after exceeding capacity")
		}
	})

	t.Run("Must not allow burst more than capacity", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))

		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}

		timeProvider.SetTime(1000) // Simulate time passing by 1 second
		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}

		for i := 0; i < 10; i++ {
			if limiter.Allow() {
				t.Errorf("Expected request %d to be not allowed", i)
			}
		}
	})

	t.Run("Must wait until tokens are available", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))

		for i := 0; i < 10; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}

		go func() {
			timeProvider.SetTime(500)
			timeProvider.SetTime(1000)
		}()
		for limiter.Wait() != nil {

		}
		if timeProvider.Now() != 1000 {
			t.Errorf("Expected time to be 1000, got: %d", timeProvider.Now())
		}
	})

	t.Run("Wait must not blocking if tokens are available", func(t *testing.T) {
		timeProvider.SetTime(0)
		limiter := ratelimiter.NewTokenBucketRateLimiter(10, 10, ratelimiter.WithTimeProvider(timeProvider))

		for i := 0; i < 5; i++ {
			if !limiter.Allow() {
				t.Errorf("Expected request %d to be allowed", i)
			}
		}

		go func() {
			<- time.After(200 * time.Millisecond)
			timeProvider.SetTime(500)
			timeProvider.SetTime(1000)
		}()

		if err := limiter.Wait(); err != nil {
			t.Errorf("Expected Wait to return immediately, got error: %v", err)
		}

		if timeProvider.Now() != 0 {
			t.Errorf("Expected time to be 0, got: %d", timeProvider.Now())
		}
	})
}
