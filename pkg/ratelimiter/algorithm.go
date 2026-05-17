package ratelimiter

// Algorithm is the strategy interface for rate limiting.
// Each implementation manages its own internal bucket state keyed by value.
type Algorithm interface {
	Allow(key string) bool
	Reset(key string)
}
