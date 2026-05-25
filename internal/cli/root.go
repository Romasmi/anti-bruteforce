package cli

import "github.com/spf13/cobra"

func NewRootCmd(serve func(configFile string) error) *cobra.Command {
	var configFile string

	root := &cobra.Command{
		Use:          "anti-bruteforce",
		Short:        "Anti-bruteforce rate limiter service",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return serve(configFile)
		},
	}

	root.Flags().StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file")

	root.AddCommand(
		newVersionCmd(),
		newClearCmd(),
		newBlacklistCmd(),
		newWhitelistCmd(),
	)

	return root
}
