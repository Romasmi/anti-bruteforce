package leackybucket

import (
	"sync"
	"time"
)

type bucket struct {
	level       float64
	lastUpdated time.Time
}

type LeakyBucketParams struct {
	Capacity float64
	LeakRate float64       // tokens per second; typically Capacity / WindowSeconds
	TTL      time.Duration // how long before an idle bucket is evicted
	// Repo is optional. When non-nil, white/black list checks are applied before bucket logic.
	// Use NewMemoryRepository for an in-memory default.
	Repo Repository
}

type LeakyBucket struct {
	capacity float64
	leakRate float64
	ttl      time.Duration
	repo     Repository
	mu       sync.Mutex
	buckets  map[string]*bucket
}

func NewLeakyBucket(params LeakyBucketParams) *LeakyBucket {
	lb := &LeakyBucket{
		capacity: params.Capacity,
		leakRate: params.LeakRate,
		ttl:      params.TTL,
		repo:     params.Repo,
		buckets:  make(map[string]*bucket),
	}
	go lb.evict()
	return lb
}

func (lb *LeakyBucket) Allow(key string) bool {
	if lb.repo != nil {
		if lb.repo.ExistsInWhiteList(key) {
			return true
		}
		if lb.repo.ExistsInBlackList(key) {
			return false
		}
	}

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

func (lb *LeakyBucket) Reset(key string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	delete(lb.buckets, key)
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
