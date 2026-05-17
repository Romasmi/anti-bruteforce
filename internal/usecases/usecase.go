package usecases

import (
	"context"

	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter"
)

type Usecase interface {
	Do(ctx context.Context, req any) (any, error)
}

func NewUsecases(logger Logger, limiter ratelimiter.RateLimiter) map[Type]Usecase {
	return map[Type]Usecase{
		Healthcheck: &HealthcheckUsecase{},
		Hello:       &HelloUsecase{Logger: logger},
		CheckAuth:   &CheckAuthUsecase{Limiter: limiter},
		ClearRate:   &ClearRateUsecase{Limiter: limiter},
	}
}
