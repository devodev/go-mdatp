package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	loggerOutput  = os.Stderr
	defaultOutput = os.Stdout

	timeFormats = []string{
		"2006-01-02",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}
)

// Execute executes the root command.
func Execute() {
	rootCmd := newCommandRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(defaultOutput, err)
	}
}

func writeOut(s string) {
	fmt.Fprintln(defaultOutput, s)
}

func newCommandRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "go-mdatp",
		Short:   "Interact with the Microsoft Defender ATP REST API.",
		Version: "0.1.0",
	}
	cmd.AddCommand(
		newCommandGenDoc(),
		newCommandFetch(),
	)
	return cmd
}

func initConfig(cfgFile string) (*Config, error) {
	viperInstance := viper.New()
	if cfgFile != "" {
		viperInstance.SetConfigFile(cfgFile)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		hd, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		viperInstance.AddConfigPath(wd)
		viperInstance.AddConfigPath(hd)
		viperInstance.SetConfigName(".go-mdatp")
		viperInstance.SetConfigType("yaml")
	}

	viperInstance.AutomaticEnv()

	err := viperInstance.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	if err := viperInstance.UnmarshalExact(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Config stores credentials and application
// specific attributes.
type Config struct {
	Global struct {
		Identifier string
	}
	Credentials struct {
		ClientID     string
		ClientSecret string
		TenantID     string
	}
}

func parseDate(param string) (time.Time, error) {
	for _, format := range timeFormats {
		parsed, err := time.Parse(format, param)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("bad datetime format")
}
