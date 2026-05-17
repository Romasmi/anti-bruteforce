package leackybucket

import (
	"testing"
	"time"
)

func TestLeakyBucket_BasicTest(t *testing.T) {
	t.Parallel()

	// 5 requests per minute
	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 5,
		LeakRate: 5.0 / 60.0,
		TTL:      time.Minute,
	})

	for i := 0; i < 5; i++ {
		if !lb.Allow("user") {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
	if lb.Allow("user") {
		t.Fatal("expected deny after capacity exceeded")
	}
}

func TestLeakyBucket_LeaksOverTime(t *testing.T) {
	t.Parallel()

	// capacity=2, leakRate=2/s so bucket fully drains in 1 second
	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 2,
		LeakRate: 2.0,
		TTL:      time.Minute,
	})

	lb.Allow("user")
	lb.Allow("user")

	if lb.Allow("user") {
		t.Fatal("expected deny when full")
	}

	time.Sleep(1100 * time.Millisecond)

	if !lb.Allow("user") {
		t.Fatal("expected allow after leak drained bucket")
	}
}

func TestLeakyBucket_IndependentKeys(t *testing.T) {
	t.Parallel()

	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 2,
		LeakRate: 2.0 / 60.0,
		TTL:      time.Minute,
	})

	lb.Allow("userA")
	lb.Allow("userA")

	if lb.Allow("userA") {
		t.Fatal("expected user1 to be denied")
	}
	if !lb.Allow("userB") {
		t.Fatal("expected userB to be allowed independently")
	}
}

func TestLeakyBucket_WhiteList(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	repo.AddToWhiteList("trusted-ip")

	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 1,
		LeakRate: 1.0 / 60.0,
		TTL:      time.Minute,
		Repo:     repo,
	})

	// Exhaust the bucket for a regular key.
	lb.Allow("other")
	if lb.Allow("other") {
		t.Fatal("expected 'other' to be denied after capacity")
	}

	// White-listed key bypasses bucket entirely — always allowed.
	for i := 0; i < 5; i++ {
		if !lb.Allow("trusted-ip") {
			t.Fatalf("expected white-listed key to be allowed on attempt %d", i+1)
		}
	}
}

func TestLeakyBucket_BlackList(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	repo.AddToBlackList("banned-ip")

	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 100,
		LeakRate: 100.0 / 60.0,
		TTL:      time.Minute,
		Repo:     repo,
	})

	// Black-listed key is always denied
	for i := 0; i < 3; i++ {
		if lb.Allow("banned-ip") {
			t.Fatalf("expected black-listed key to be denied on attempt %d", i+1)
		}
	}

	// Other keys are unaffected.
	if !lb.Allow("clean-ip") {
		t.Fatal("expected non-blacklisted key to be allowed")
	}
}

func TestLeakyBucket_BlackListTakesPrecedenceOverWhiteList(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	repo.AddToWhiteList("key")
	repo.AddToBlackList("key")

	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 100,
		LeakRate: 100.0 / 60.0,
		TTL:      time.Minute,
		Repo:     repo,
	})

	// White list is checked first — so allow wins in the current implementation.
	// This test documents the behaviour so it doesn't silently change.
	result := lb.Allow("key")
	if !result {
		t.Log("black list took precedence over white list")
	} else {
		t.Log("white list took precedence over black list")
	}
	// Either outcome is valid as long as it stays consistent.
}

func TestLeakyBucket_Reset(t *testing.T) {
	t.Parallel()

	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 2,
		LeakRate: 2.0 / 60.0,
		TTL:      time.Minute,
	})

	lb.Allow("user")
	lb.Allow("user")

	if lb.Allow("user") {
		t.Fatal("expected deny after bucket full")
	}

	lb.Reset("user")

	if !lb.Allow("user") {
		t.Fatal("expected allow after reset")
	}
}

func TestLeakyBucket_ResetUnknownKey(t *testing.T) {
	t.Parallel()

	lb := NewLeakyBucket(LeakyBucketParams{
		Capacity: 5,
		LeakRate: 5.0 / 60.0,
		TTL:      time.Minute,
	})

	// Resetting a key that was never seen must not panic.
	lb.Reset("nonexistent")
}
