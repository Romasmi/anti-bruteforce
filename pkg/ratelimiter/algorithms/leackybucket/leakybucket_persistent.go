package leackybucket

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// allowScript atomically applies leaky-bucket logic in Redis.
// It reads the bucket state, drains it by the leaked amount, and increments if capacity allows.
// Returns 1 if the request is allowed, 0 if denied.
var allowScript = redis.NewScript(`
local key     = KEYS[1]
local cap     = tonumber(ARGV[1])
local rate    = tonumber(ARGV[2])
local now_ms  = tonumber(ARGV[3])
local ttl_ms  = tonumber(ARGV[4])

local data = redis.call('HMGET', key, 'level', 'ts')
local level = tonumber(data[1]) or 0
local ts    = tonumber(data[2]) or now_ms

local elapsed = math.max(0.0, (now_ms - ts) / 1000.0)
level = math.max(0.0, level - elapsed * rate)

if level + 1 > cap then
    return 0
end

level = level + 1
redis.call('HSET', key, 'level', tostring(level), 'ts', tostring(now_ms))
redis.call('PEXPIRE', key, ttl_ms)
return 1
`)

// PersistentLeakyBucket is a Redis-backed leaky bucket.
// Its state is shared across all app instances, so Reset works globally.
type PersistentLeakyBucket struct {
	client   *redis.Client
	prefix   string
	capacity float64
	leakRate float64
	ttl      time.Duration
	repo     Repository
}

type PersistentLeakyBucketParams struct {
	Client   *redis.Client
	Prefix   string // namespace prefix, e.g. "login", "ip"
	Capacity float64
	LeakRate float64
	TTL      time.Duration
	Repo     Repository
}

func NewPersistentLeakyBucket(params PersistentLeakyBucketParams) *PersistentLeakyBucket {
	return &PersistentLeakyBucket{
		client:   params.Client,
		prefix:   params.Prefix,
		capacity: params.Capacity,
		leakRate: params.LeakRate,
		ttl:      params.TTL,
		repo:     params.Repo,
	}
}

func (lb *PersistentLeakyBucket) Allow(key string) bool {
	if lb.repo != nil {
		if lb.repo.ExistsInWhiteList(key) {
			return true
		}
		if lb.repo.ExistsInBlackList(key) {
			return false
		}
	}

	ctx := context.Background()
	rk := lb.redisKey(key)
	nowMs := time.Now().UnixMilli()
	ttlMs := lb.ttl.Milliseconds()

	res, err := allowScript.Run(ctx, lb.client, []string{rk},
		lb.capacity, lb.leakRate, nowMs, ttlMs,
	).Int()
	if err != nil {
		// Fail open on Redis error to avoid blocking legitimate traffic.
		return true
	}
	return res == 1
}

func (lb *PersistentLeakyBucket) Reset(key string) {
	lb.client.Del(context.Background(), lb.redisKey(key))
}

func (lb *PersistentLeakyBucket) redisKey(key string) string {
	return fmt.Sprintf("rl_lb:%s:%s", lb.prefix, key)
}
