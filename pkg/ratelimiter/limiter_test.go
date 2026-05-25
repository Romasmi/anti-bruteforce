package ratelimiter

import (
	"testing"
	"time"

	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter/algorithms/leackybucket"
)

const (
	StrategyLogin    = "login"
	StrategyPassword = "password"
	StrategyIP       = "ip"
)

func TestRateLimiter_LoginStrategy_BlocksAfterCapacity(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(AlgorithmMap{StrategyLogin: bucket(3)})

	for i := 0; i < 3; i++ {
		ok, err := rl.Allow(StrategyMap{StrategyLogin: "userA"})
		if err != nil || !ok {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	ok, err := rl.Allow(StrategyMap{StrategyLogin: "userA"})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected denial after login bucket exhausted")
	}
}

func TestRateLimiter_PasswordStrategy_BlocksAfterCapacity(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(AlgorithmMap{StrategyPassword: bucket(3)})

	for i := 0; i < 3; i++ {
		ok, err := rl.Allow(StrategyMap{StrategyPassword: "userpass"})
		if err != nil || !ok {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	ok, err := rl.Allow(StrategyMap{StrategyPassword: "userpass"})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected denial after password bucket exhausted")
	}
}

func TestRateLimiter_IPStrategy_BlocksAfterCapacity(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(AlgorithmMap{StrategyIP: bucket(3)})

	for i := 0; i < 3; i++ {
		ok, err := rl.Allow(StrategyMap{StrategyIP: "127.0.0.1"})
		if err != nil || !ok {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	ok, err := rl.Allow(StrategyMap{StrategyIP: "127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected denial after ip bucket exhausted")
	}
}

// TestRateLimiter_DenialWhenAnyStrategyFails verifies that a single exhausted
// strategy denies the request even when all other strategies still have room.
func TestRateLimiter_DenialWhenAnyStrategyFails(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(AlgorithmMap{
		StrategyLogin:    bucket(1),
		StrategyPassword: bucket(100),
		StrategyIP:       bucket(1000),
	})
	checks := StrategyMap{
		StrategyLogin:    "userA",
		StrategyPassword: "userpass",
		StrategyIP:       "127.0.0.1",
	}

	ok, err := rl.Allow(checks)
	if err != nil || !ok {
		t.Fatal("expected first request to be allowed")
	}

	// login bucket exhausted; password and ip still have room
	ok, err = rl.Allow(checks)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected denial when one strategy bucket is exhausted")
	}
}

func TestRateLimiter_UnknownStrategy(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(AlgorithmMap{StrategyLogin: bucket(10)})

	_, err := rl.Allow(StrategyMap{"unknown": "value"})
	if err == nil {
		t.Fatal("expected error for unknown strategy")
	}
}

func bucket(capacity float64) *leackybucket.LeakyBucket {
	return leackybucket.NewLeakyBucket(leackybucket.LeakyBucketParams{
		Capacity: capacity,
		LeakRate: capacity / 60.0,
		TTL:      time.Minute,
	})
}
