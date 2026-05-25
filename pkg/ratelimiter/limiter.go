package ratelimiter

import (
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type (
	Strategy     = string
	AlgorithmMap = map[string]Algorithm
	StrategyMap  = map[string]string
)

type Impl struct {
	strategies AlgorithmMap
}

type RateLimiter interface {
	Allow(checks map[string]string) (bool, error)
	Reset(strategy, key string) error
}

func NewRateLimiter(strategies AlgorithmMap) RateLimiter {
	return &Impl{strategies: strategies}
}

func (rl *Impl) Allow(checks StrategyMap) (bool, error) {
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

func (rl *Impl) Reset(strategy, key string) error {
	alg, ok := rl.strategies[strategy]
	if !ok {
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
	alg.Reset(key)
	return nil
}
