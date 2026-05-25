package cli

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

var (
	release   = "UNKNOWN"
	buildDate = "UNKNOWN"
	gitHash   = "UNKNOWN"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Print version info",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return json.NewEncoder(os.Stdout).Encode(struct {
				Release   string
				BuildDate string
				GitHash   string
			}{
				Release:   release,
				BuildDate: buildDate,
				GitHash:   gitHash,
			})
		},
	}
}
