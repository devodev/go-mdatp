package cmd

import (
	"context"
	"encoding/json"
	"go-mdatp/pkg/mdatp"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

type fetchConfig struct {
	ConfigFile string
}

// setupCmd sets flags on the provided cmd and resolve env variables using the provided Config.
func fetchSetupCmd(cmd *cobra.Command, c *fetchConfig) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(&c.ConfigFile, "config", "c", c.ConfigFile, "config file (default is $CWD/.go-mdatp.yaml)")
	return cmd
}

func newCommandFetch() *cobra.Command {
	var cmdConfig fetchConfig
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch alerts from the Microsoft Defender ATP REST API.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := initConfig(cmdConfig.ConfigFile)
			if err != nil {
				return err
			}

			client, err := mdatp.NewClient(
				mdatp.WithOAuthClient(
					config.Credentials.ClientID,
					config.Credentials.ClientSecret,
					config.Credentials.TenantID,
				),
			)
			if err != nil {
				return err
			}

			resp, alert, err := client.Alert.Fetch(context.Background())
			if err != nil {
				return err
			}

			if resp.APIError != nil {
				marshalled, err := json.Marshal(resp.APIError)
				if err != nil {
					return err
				}
				writeOut(string(marshalled))
				return nil
			}

			for _, a := range alert.Value {
				marshalled, err := json.Marshal(a)
				if err != nil {
					return err
				}
				writeOut(string(marshalled))
			}
			return nil
		},
	}
	return fetchSetupCmd(cmd, &cmdConfig)
}
