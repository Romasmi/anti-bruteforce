package usecases

import (
	"context"

	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CheckAuthUsecase struct {
	Limiter ratelimiter.RateLimiter
}

func (u *CheckAuthUsecase) Do(_ context.Context, req any) (any, error) {
	r, ok := req.(*CheckAuthInput)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	allowed, err := u.Limiter.Allow(map[string]string{
		StrategyLogin:    r.Login,
		StrategyPassword: r.Password,
		StrategyIP:       r.IP,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rate limiter: %v", err)
	}

	return allowed, nil
}
