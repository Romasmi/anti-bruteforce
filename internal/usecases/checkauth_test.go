package usecases_test

import (
	"testing"
	"time"

	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter"
	leackybucket "github.com/Romasmi/anti-bruteforce/pkg/ratelimiter/algorithms/leackybucket"
	"github.com/stretchr/testify/require"
)

func TestCheckAuthUsecase_Validation(t *testing.T) {
	t.Parallel()
	uc := &usecases.CheckAuthUsecase{Limiter: allowAllLimiter{}}
	ctx := t.Context()

	t.Run("empty login", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Login: "", Password: "password123", IP: "192.168.1.1"})
		requireCode(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Login: "user123", Password: "", IP: "192.168.1.1"})
		requireCode(t, err)
	})

	t.Run("empty ip", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Login: "user123", Password: "password123", IP: ""})
		requireCode(t, err)
	})

	t.Run("invalid ip format", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Login: "user123", Password: "password123", IP: "not-an-ip"})
		requireCode(t, err)
	})

	t.Run("ipv6 rejected", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.CheckAuthInput{Login: "user123", Password: "password123", IP: "2001:db8::1"})
		requireCode(t, err)
	})

	t.Run("wrong request type", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, "not a CheckAuthInput")
		requireCode(t, err)
	})
}

func TestCheckAuthUsecase_Allowed(t *testing.T) {
	t.Parallel()
	uc := &usecases.CheckAuthUsecase{Limiter: allowAllLimiter{}}

	res, err := uc.Do(t.Context(), &usecases.CheckAuthInput{
		Login:    "user123",
		Password: "password123",
		IP:       "192.168.1.1",
	})
	require.NoError(t, err)
	require.True(t, res.(bool))
}

func TestCheckAuthUsecase_Denied(t *testing.T) {
	t.Parallel()
	uc := &usecases.CheckAuthUsecase{Limiter: denyAllLimiter{}}

	res, err := uc.Do(t.Context(), &usecases.CheckAuthInput{
		Login:    "user123",
		Password: "password123",
		IP:       "192.168.1.1",
	})
	require.NoError(t, err)
	require.False(t, res.(bool))
}

// TestCheckAuthUsecase_RateLimiting verifies that CheckAuth returns ok=false once
// a bucket is exhausted, using a real leaky-bucket limiter.
func TestCheckAuthUsecase_RateLimiting(t *testing.T) {
	t.Parallel()

	limiter := ratelimiter.NewRateLimiter(ratelimiter.AlgorithmMap{
		usecases.StrategyLogin: leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
			Capacity: 2,
			LeakRate: 2.0 / 60.0,
			TTL:      time.Minute,
		}),
		usecases.StrategyPassword: leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
			Capacity: 100,
			LeakRate: 100.0 / 60.0,
			TTL:      time.Minute,
		}),
		usecases.StrategyIP: leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
			Capacity: 1000,
			LeakRate: 1000.0 / 60.0,
			TTL:      time.Minute,
		}),
	})
	uc := &usecases.CheckAuthUsecase{Limiter: limiter}
	ctx := t.Context()
	input := &usecases.CheckAuthInput{Login: "alice", Password: "s3cr3t", IP: "10.0.0.1"}

	for i := 0; i < 2; i++ {
		res, err := uc.Do(ctx, input)
		require.NoError(t, err)
		require.True(t, res.(bool), "expected request %d to be allowed", i+1)
	}

	res, err := uc.Do(ctx, input)
	require.NoError(t, err)
	require.False(t, res.(bool), "expected denial after login bucket exhausted")
}
