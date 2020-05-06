package cmd

import (
	"context"
	"encoding/json"
	"go-mdatp/pkg/mdatp"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var (
	config configAlert
)

type configAlert struct {
	ConfigFile string
}

// setupCmd sets flags on the provided cmd and resolve env variables using the provided Config.
func setupCmdAlert(cmd *cobra.Command, c *configAlert) *cobra.Command {
	envconfig.Process("", c)
	cmd.PersistentFlags().SortFlags = false
	cmd.PersistentFlags().StringVarP(&c.ConfigFile, "config", "c", c.ConfigFile, "config file (default is $CWD/.go-mdatp.yaml)")
	return cmd
}

func newCommandAlert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Alert resource type commands.",
	}
	cmd.AddCommand(
		newCommandAlertList(),
	)
	return setupCmdAlert(cmd, &config)
}

func newCommandAlertList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List alerts.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := initConfig(config.ConfigFile)
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

			resp, alert, err := client.Alert.List(context.Background())
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
	return cmd
}
