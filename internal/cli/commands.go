package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultAddr    = "localhost:50051"
	defaultTimeout = 5 * time.Second
)

func newClearCmd() *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:          "clear <login|ip> <value>",
		Short:        "Clear rate limit bucket for a login or IP",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context())
			defer cancel()
			return ClearRate(ctx, addr, args[0], args[1])
		},
	}

	cmd.Flags().StringVar(&addr, "addr", defaultAddr, "anti-bruteforce gRPC address (host:port)")
	return cmd
}

func newBlacklistCmd() *cobra.Command {
	return newIPListCmd("blacklist")
}

func newWhitelistCmd() *cobra.Command {
	return newIPListCmd("whitelist")
}

func newIPListCmd(list string) *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:          list + " <add|remove> <subnet>",
		Short:        "Manage " + list + " entries",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context())
			defer cancel()
			return UpdateIPList(ctx, addr, list, args[0], args[1])
		},
	}

	cmd.Flags().StringVar(&addr, "addr", defaultAddr, "anti-bruteforce gRPC address (host:port)")
	return cmd
}

func withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, defaultTimeout)
}
