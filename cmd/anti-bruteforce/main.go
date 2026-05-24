package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Romasmi/anti-bruteforce/internal/app/api"
	"github.com/Romasmi/anti-bruteforce/internal/logger"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	switch flag.Arg(0) {
	case "version":
		printVersion()
		return
	case "clear":
		if err := runClear(flag.Args()[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	case "blacklist":
		if err := runBlacklist(flag.Args()[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	case "whitelist":
		if err := runWhitelist(flag.Args()[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
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

	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("failed to run app: %w", err)
	}

	return nil
}
