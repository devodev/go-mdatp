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
		newCommandWatch(),
	)
	return setupCmdAlert(cmd, &config)
}

type configAlertList struct {
	ODataQueryFilter string
}

func setupCmdAlertList(cmd *cobra.Command, c *configAlertList) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(&c.ODataQueryFilter, "query-filter", "f", c.ODataQueryFilter, "$filter OData V4 query option string.")
	return cmd
}

func newCommandAlertList() *cobra.Command {
	var cmdConfig configAlertList
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

			resp, alert, err := client.Alert.List(context.Background(), cmdConfig.ODataQueryFilter)
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
	return setupCmdAlertList(cmd, &cmdConfig)
}
