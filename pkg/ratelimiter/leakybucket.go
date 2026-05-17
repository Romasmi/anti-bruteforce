package ratelimiter

import (
	"sync"
	"time"
)

type bucket struct {
	level       float64
	lastUpdated time.Time
}

type LeakyBucketConfig struct {
	Capacity float64
	LeakRate float64       // tokens per second; typically Capacity / WindowSeconds
	TTL      time.Duration // how long before an idle bucket is evicted
}

type LeakyBucket struct {
	capacity float64
	leakRate float64
	ttl      time.Duration
	mu       sync.Mutex
	buckets  map[string]*bucket
}

func NewLeakyBucket(config LeakyBucketConfig) *LeakyBucket {
	lb := &LeakyBucket{
		capacity: config.Capacity,
		leakRate: config.LeakRate,
		ttl:      config.TTL,
		buckets:  make(map[string]*bucket),
	}
	go lb.evict()
	return lb
}

func (lb *LeakyBucket) Allow(key string) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	b, ok := lb.buckets[key]
	if !ok {
		b = &bucket{lastUpdated: now}
		lb.buckets[key] = b
	}

	elapsed := now.Sub(b.lastUpdated).Seconds()
	b.level = max(0, b.level-elapsed*lb.leakRate)
	b.lastUpdated = now

	if b.level+1 > lb.capacity {
		return false
	}

	b.level++
	return true
}

func (lb *LeakyBucket) evict() {
	ticker := time.NewTicker(lb.ttl)
	defer ticker.Stop()
	for range ticker.C {
		lb.mu.Lock()
		for key, b := range lb.buckets {
			if time.Since(b.lastUpdated) > lb.ttl {
				delete(lb.buckets, key)
			}
		}
		lb.mu.Unlock()
	}
}
