package usecases

import (
	"context"

	"github.com/Romasmi/anti-bruteforce/pkg/ratelimiter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClearRateUsecase struct {
	Limiter ratelimiter.RateLimiter
}

func (u *ClearRateUsecase) Do(_ context.Context, req any) (any, error) {
	r, ok := req.(*ClearRateInput)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	if r.Login != "" {
		if err := u.Limiter.Reset(StrategyLogin, r.Login); err != nil {
			return nil, status.Errorf(codes.Internal, "reset login bucket: %v", err)
		}
	}

	if r.IP != "" {
		if err := u.Limiter.Reset(StrategyIP, r.IP); err != nil {
			return nil, status.Errorf(codes.Internal, "reset ip bucket: %v", err)
		}
	}

	return nil, nil
}
