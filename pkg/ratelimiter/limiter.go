package ratelimiter

import (
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type RateLimiter struct {
	strategies map[Strategy]Algorithm
}

func NewRateLimiter(strategies map[Strategy]Algorithm) *RateLimiter {
	return &RateLimiter{strategies: strategies}
}

func (rl *RateLimiter) Allow(checks StrategyMap) (bool, error) {
	for strategy := range checks {
		if _, ok := rl.strategies[strategy]; !ok {
			return false, fmt.Errorf("unknown strategy: %s", strategy)
		}
	}

	var (
		g      errgroup.Group
		denied atomic.Bool
	)

	for strategy, value := range checks {
		alg := rl.strategies[strategy]
		v := value
		g.Go(func() error {
			if !alg.Allow(v) {
				denied.Store(true)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}
	return !denied.Load(), nil
}
