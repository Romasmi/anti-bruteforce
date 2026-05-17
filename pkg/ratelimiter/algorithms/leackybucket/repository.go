package leackybucket

import "sync"

// Repository is the read-only interface LeakyBucket uses to check membership lists.
// Implement it with any backing store (Redis, DB, etc.) and pass it via LeakyBucketParams.
type Repository interface {
	ExistsInWhiteList(key string) bool
	ExistsInBlackList(key string) bool
}

// MemoryRepository is a thread-safe in-memory Repository.
type MemoryRepository struct {
	mu        sync.RWMutex
	whiteList map[string]struct{}
	blackList map[string]struct{}
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		whiteList: make(map[string]struct{}),
		blackList: make(map[string]struct{}),
	}
}

func (r *MemoryRepository) AddToWhiteList(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.whiteList[key] = struct{}{}
}

func (r *MemoryRepository) RemoveFromWhiteList(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.whiteList, key)
}

func (r *MemoryRepository) AddToBlackList(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blackList[key] = struct{}{}
}

func (r *MemoryRepository) RemoveFromBlackList(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.blackList, key)
}

func (r *MemoryRepository) ExistsInWhiteList(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.whiteList[key]
	return ok
}

func (r *MemoryRepository) ExistsInBlackList(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.blackList[key]
	return ok
}
