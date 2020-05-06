package cmd

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

type fetchConfig struct {
}

// setupCmd sets flags on the provided cmd and resolve env variables using the provided Config.
func fetchSetupCmd(cmd *cobra.Command, c *fetchConfig) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().SortFlags = false
	return cmd
}

func newCommandFetch() *cobra.Command {
	var config fetchConfig
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch alerts from the Microsoft Defender ATP SIEM API.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return fetchSetupCmd(cmd, &config)
}
