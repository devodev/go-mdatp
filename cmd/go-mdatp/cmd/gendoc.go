package cmd

import (
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type genDocConfig struct {
	OutDir string `envconfig:"OUT_DIR" default:"./docs"`
}

// setupCmd sets flags on the provided cmd and resolve env variables using the provided Config.
func setupCmd(cmd *cobra.Command, c *genDocConfig) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&c.OutDir, "dir", c.OutDir, "directory where to write the doc.")
	return cmd
}

func newCommandGenDoc() *cobra.Command {
	var config genDocConfig
	cmd := &cobra.Command{
		Use:   "gendoc",
		Short: "Generate markdown documentation for the go-mdatp CLI.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fpath, err := filepath.Abs(config.OutDir)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}

			if err := doc.GenMarkdownTree(newCommandRoot(), fpath); err != nil {
				return err
			}
			return nil
		},
	}
	return setupCmd(cmd, &config)
}
