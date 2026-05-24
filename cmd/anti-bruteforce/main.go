package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Romasmi/anti-bruteforce/internal/app/api"
	"github.com/Romasmi/anti-bruteforce/internal/cli"
	"github.com/Romasmi/anti-bruteforce/internal/logger"
)

func main() {
	if err := cli.NewRootCmd(runServer).Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(configFile string) error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	conf, err := api.NewConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	l := logger.New(conf.Logger.Level, "anti-bruteforce")

	app := api.New(conf, l)

	if err := app.Init(ctx); err != nil {
		return fmt.Errorf("failed to init app: %w", err)
	}

	return app.Run(ctx)
}
