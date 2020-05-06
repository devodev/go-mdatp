package cmd

import (
	"context"
	"encoding/json"
	"go-mdatp/pkg/mdatp"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var (
	sinceTimeUTCHelp             = "Defines the lower time bound alerts are retrieved from, based on alert field: LastProcessedTimeUtc."
	untilTimeUTCHelp             = "Defines the upper time bound alerts are retrieved. The time range will be: from sinceTimeUtc time to untilTimeUtc time."
	agoHelp                      = "A ISO8601 duration forma string used to pull alerts in the following time range: from (current_time - ago) time to current_time time."
	limitHelp                    = "The number of alerts to be retrieved. Most recent alerts will be retrieved based on the number defined."
	machineGroupsHelp            = "A comma separated list of machine groups to pull alerts from."
	deviceCreatedMachineTagsHelp = "A single machine tag from the registry."
	cloudCreatedMachineTagsHelp  = "A comma separated list of machine tags that were created in Microsoft Defender Security Center."
)

type fetchConfig struct {
	ConfigFile               string
	SinceTimeUTC             string
	UntilTimeUTC             string
	Ago                      string
	Limit                    int
	Machinegroups            []string
	DeviceCreatedMachineTags string
	CloudCreatedMachineTags  []string
}

func (c *fetchConfig) toAlertRequestParams() (*mdatp.AlertRequestParams, error) {
	var err error
	var since time.Time
	var until time.Time
	if c.SinceTimeUTC != "" {
		since, err = parseDate(c.SinceTimeUTC)
		if err != nil {
			return nil, err
		}
	}
	if c.UntilTimeUTC != "" {
		until, err = parseDate(c.UntilTimeUTC)
		if err != nil {
			return nil, err
		}
	}
	return &mdatp.AlertRequestParams{
		SinceTimeUTC:             since,
		UntilTimeUTC:             until,
		Ago:                      c.Ago,
		Limit:                    c.Limit,
		Machinegroups:            c.Machinegroups,
		DeviceCreatedMachineTags: c.DeviceCreatedMachineTags,
		CloudCreatedMachineTags:  c.CloudCreatedMachineTags,
	}, nil
}

// setupCmd sets flags on the provided cmd and resolve env variables using the provided Config.
func fetchSetupCmd(cmd *cobra.Command, c *fetchConfig) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(&c.ConfigFile, "config", "c", c.ConfigFile, "config file (default is $CWD/.go-mdatp.yaml)")
	cmd.Flags().StringVarP(&c.SinceTimeUTC, "since-time-utc", "s", c.SinceTimeUTC, sinceTimeUTCHelp)
	cmd.Flags().StringVarP(&c.UntilTimeUTC, "until-time-utc", "u", c.UntilTimeUTC, untilTimeUTCHelp)
	cmd.Flags().StringVarP(&c.Ago, "ago", "a", c.Ago, agoHelp)
	cmd.Flags().IntVarP(&c.Limit, "limit", "l", c.Limit, limitHelp)
	cmd.Flags().StringSliceVarP(&c.Machinegroups, "machine-groups", "m", c.Machinegroups, machineGroupsHelp)
	cmd.Flags().StringVar(&c.DeviceCreatedMachineTags, "device-created-machine-tags", c.DeviceCreatedMachineTags, deviceCreatedMachineTagsHelp)
	cmd.Flags().StringSliceVar(&c.CloudCreatedMachineTags, "cloud-created-machine-tags", c.CloudCreatedMachineTags, cloudCreatedMachineTagsHelp)
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
			requestParams, err := cmdConfig.toAlertRequestParams()
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

			resp, alert, err := client.Alert.Fetch(context.Background(), requestParams)
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
