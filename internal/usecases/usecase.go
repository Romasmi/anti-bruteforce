package usecases

import (
	"context"
)

type Usecase interface {
	Do(ctx context.Context, req any) (any, error)
}

func NewUsecases(logger Logger) map[Type]Usecase {
	return map[Type]Usecase{
		Healthcheck: &HealthcheckUsecase{},
		Hello:       &HelloUsecase{Logger: logger},
	}
}
