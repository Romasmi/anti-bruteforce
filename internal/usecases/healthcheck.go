package usecases

import (
	"context"
)

type HealthcheckUsecase struct{}

func (u *HealthcheckUsecase) Do(_ context.Context, _ any) (any, error) {
	return "ok", nil
}
