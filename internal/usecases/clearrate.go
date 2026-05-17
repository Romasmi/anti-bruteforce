package usecases

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClearRateUsecase struct{}

func (u *ClearRateUsecase) Do(_ context.Context, req any) (any, error) {
	r, ok := req.(*ClearRateInput)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request type")
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	return nil, nil
}
