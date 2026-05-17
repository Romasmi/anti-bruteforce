package leackybucket

import (
	"testing"
	"time"
)

func TestLeakyBucket_BasicTest(t *testing.T) {
	// 5 requests per minute
	lb := NewLeakyBucket(LeakyBucketConfig{
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
	// capacity=2, leakRate=2/s so bucket fully drains in 1 second
	lb := NewLeakyBucket(LeakyBucketConfig{
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
	lb := NewLeakyBucket(LeakyBucketConfig{
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
