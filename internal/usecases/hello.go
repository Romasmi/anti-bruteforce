package usecases

import (
	"context"
	"fmt"
	"time"
)

type Logger interface {
	Info(msg string)
}

type HelloUsecase struct {
	Logger Logger
}

func (u *HelloUsecase) Do(_ context.Context, req any) (any, error) {
	u.Logger.Info(fmt.Sprintf("request: %v, time: %s", req, time.Now().Format(time.RFC3339)))
	return "hello world", nil
}
