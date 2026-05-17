package usecases_test

import (
	"testing"
	"time"

	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter"
	leackybucket "github.com/Romasmi/anti-bruteforce/pkg/ratelimiter/algorithms/leackybucket"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestClearRateUsecase_Validation(t *testing.T) {
	t.Parallel()
	uc := &usecases.ClearRateUsecase{Limiter: allowAllLimiter{}}
	ctx := t.Context()

	t.Run("empty both", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Login: "", IP: ""})
		requireCode(t, err, codes.InvalidArgument)
	})

	t.Run("invalid ip format", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.ClearRateInput{IP: "999.999.999.999"})
		requireCode(t, err, codes.InvalidArgument)
	})

	t.Run("ipv6 rejected", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, &usecases.ClearRateInput{IP: "::1"})
		requireCode(t, err, codes.InvalidArgument)
	})

	t.Run("wrong request type", func(t *testing.T) {
		t.Parallel()
		_, err := uc.Do(ctx, 42)
		requireCode(t, err, codes.InvalidArgument)
	})
}

func TestClearRateUsecase_ValidInputs(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("login only", func(t *testing.T) {
		t.Parallel()
		uc := &usecases.ClearRateUsecase{Limiter: allowAllLimiter{}}
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Login: "user123"})
		require.NoError(t, err)
	})

	t.Run("ip only", func(t *testing.T) {
		t.Parallel()
		uc := &usecases.ClearRateUsecase{Limiter: allowAllLimiter{}}
		_, err := uc.Do(ctx, &usecases.ClearRateInput{IP: "10.0.0.1"})
		require.NoError(t, err)
	})

	t.Run("both login and ip", func(t *testing.T) {
		t.Parallel()
		uc := &usecases.ClearRateUsecase{Limiter: allowAllLimiter{}}
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Login: "user123", IP: "10.0.0.1"})
		require.NoError(t, err)
	})
}

// TestClearRateUsecase_ResetsLoginBucket verifies that ClearRate restores the
// login bucket so that a previously-denied auth attempt is allowed again.
func TestClearRateUsecase_ResetsLoginBucket(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	limiter := realLimiter(1, 100, 1000)
	checkAuth := &usecases.CheckAuthUsecase{Limiter: limiter}
	clearRate := &usecases.ClearRateUsecase{Limiter: limiter}
	input := &usecases.CheckAuthInput{Login: "alice", Password: "s3cr3t", IP: "10.0.0.1"}

	res, err := checkAuth.Do(ctx, input)
	require.NoError(t, err)
	require.True(t, res.(bool), "first request should be allowed")

	res, err = checkAuth.Do(ctx, input)
	require.NoError(t, err)
	require.False(t, res.(bool), "second request should be denied (login bucket full)")

	_, err = clearRate.Do(ctx, &usecases.ClearRateInput{Login: "alice"})
	require.NoError(t, err)

	res, err = checkAuth.Do(ctx, input)
	require.NoError(t, err)
	require.True(t, res.(bool), "request after reset should be allowed again")
}

// TestClearRateUsecase_ResetsIPBucket verifies that ClearRate restores the
// IP bucket so that a previously-denied auth attempt is allowed again.
func TestClearRateUsecase_ResetsIPBucket(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	limiter := realLimiter(100, 100, 1)
	checkAuth := &usecases.CheckAuthUsecase{Limiter: limiter}
	clearRate := &usecases.ClearRateUsecase{Limiter: limiter}
	input := &usecases.CheckAuthInput{Login: "bob", Password: "p4ss", IP: "10.0.0.2"}

	res, err := checkAuth.Do(ctx, input)
	require.NoError(t, err)
	require.True(t, res.(bool), "first request should be allowed")

	res, err = checkAuth.Do(ctx, input)
	require.NoError(t, err)
	require.False(t, res.(bool), "second request should be denied (ip bucket full)")

	_, err = clearRate.Do(ctx, &usecases.ClearRateInput{IP: "10.0.0.2"})
	require.NoError(t, err)

	res, err = checkAuth.Do(ctx, input)
	require.NoError(t, err)
	require.True(t, res.(bool), "request after reset should be allowed again")
}

func realLimiter(loginCap, passwordCap, ipCap float64) ratelimiter.RateLimiter {
	b := func(capacity float64) *leackybucket.LeakyBucket {
		return leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
			Capacity: capacity,
			LeakRate: capacity / 60.0,
			TTL:      time.Minute,
		})
	}
	return ratelimiter.NewRateLimiter(ratelimiter.AlgorithmMap{
		usecases.StrategyLogin:    b(loginCap),
		usecases.StrategyPassword: b(passwordCap),
		usecases.StrategyIP:       b(ipCap),
	})
}
