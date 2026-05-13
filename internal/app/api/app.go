package api

import (
	"context"
	"fmt"
	"time"

	"github.com/Romasmi/anti-bruteforce/internal/logger"
	grpcserver "github.com/Romasmi/anti-bruteforce/internal/server/grpc"
	internalhttp "github.com/Romasmi/anti-bruteforce/internal/server/http"
	"github.com/Romasmi/anti-bruteforce/internal/usecases"
)

type App struct {
	config     Config
	logger     *logger.Logger
	grpcServer *grpcserver.Server
	httpServer *internalhttp.Server
}

func New(conf Config, l *logger.Logger) *App {
	return &App{
		config: conf,
		logger: l,
	}
}

func (a *App) Init(_ context.Context) error {
	ucs := usecases.NewUsecases(a.logger)

	a.grpcServer = grpcserver.NewServer(a.logger, ucs)
	grpcAddr := fmt.Sprintf("%s:%s", a.config.GRPC.Host, a.config.GRPC.Port)
	a.httpServer = internalhttp.NewServer(a.logger, grpcAddr, a.config.HTTP.Host, a.config.HTTP.Port)

	return nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		if err := a.grpcServer.Start(a.config.GRPC.Host, a.config.GRPC.Port); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	go func() {
		if err := a.httpServer.Start(ctx); err != nil {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	a.logger.Info("anti-bruteforce is running...")

	select {
	case <-ctx.Done():
		a.logger.Info("Stopping servers...")
	case err := <-errCh:
		a.logger.Error("Server error: " + err.Error())
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer stopCancel()

	if err := a.httpServer.Stop(stopCtx); err != nil {
		a.logger.Error("failed to stop http server: " + err.Error())
	}
	a.grpcServer.Stop()

	return nil
}
