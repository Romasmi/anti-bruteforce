package ratelimiter

import (
	"testing"
	"time"

	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter/algorithms/leackybucket"
)

const (
	StrategyLogin    Strategy = "login"
	StrategyPassword Strategy = "password"
	StrategyIP       Strategy = "ip"
)

func TestRateLimiter_LoginStrategy_BlocksAfterCapacity(t *testing.T) {
	t.Parallel()

	rl := newLimiter()
	checks := StrategyMap{
		StrategyLogin:    "userA",
		StrategyPassword: "secret",
		StrategyIP:       "127.0.0.1",
	}

	ok, err := rl.Allow(checks)
	if err != nil || !ok {
		t.Fatal("expected first request to be allowed")
	}

	// login bucket capacity = 1
	ok, err = rl.Allow(checks)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected denial after login bucket exhausted")
	}
}

func TestRateLimiter_PasswordStrategy_BlocksAfterCapacity(t *testing.T) {
	t.Parallel()

	rl := newLimiter()
	checks := StrategyMap{
		StrategyLogin:    "userB",
		StrategyPassword: "shared-password",
		StrategyIP:       "127.0.0.2",
	}

	// password capacity = 2
	for i := 0; i < 2; i++ {
		ok, err := rl.Allow(checks)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("expected request %d to pass", i+1)
		}
	}

	ok, err := rl.Allow(checks)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected denial after password bucket exhausted")
	}
}

func TestRateLimiter_IPStrategy_BlocksAfterCapacity(t *testing.T) {
	t.Parallel()

	rl := newLimiter()
	checks := StrategyMap{
		StrategyLogin:    "userC",
		StrategyPassword: "another-password",
		StrategyIP:       "10.0.0.1",
	}

	// ip capacity = 3
	for i := 0; i < 3; i++ {
		ok, err := rl.Allow(checks)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("expected request %d to pass", i+1)
		}
	}

	ok, err := rl.Allow(checks)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected denial after IP bucket exhausted")
	}
}

func TestRateLimiter_UnknownStrategy(t *testing.T) {
	t.Parallel()

	rl := newLimiter()

	_, err := rl.Allow(StrategyMap{
		"unknown": "value",
	})
	if err == nil {
		t.Fatal("expected error for unknown strategy")
	}
}

func newLimiter() *RateLimiter {
	return NewRateLimiter(AlgorithmMap{
		StrategyLogin: leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
			Capacity: 1,
			LeakRate: 1.0 / 60.0,
			TTL:      time.Minute,
		}),
		StrategyPassword: leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
			Capacity: 100,
			LeakRate: 100.0 / 60.0,
			TTL:      time.Minute,
		}),
		StrategyIP: leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
			Capacity: 1000,
			LeakRate: 1000.0 / 60.0,
			TTL:      2 * time.Minute,
		}),
	})
}
