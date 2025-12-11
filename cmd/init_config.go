package cmd

import (
	"os"

	"github.com/sydneyowl/clh-server/pkg/config"

	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

var overwriteFile bool

// initConfigCmd represents the initConfig command
var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Release a empty config template.",
	Long:  `Release a empty config template.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		template := config.GenerateDefaultConfig()
		data, err := yaml.Marshal(template)
		if err != nil {
			return err
		}

		if cfgFilePath == "" {
			slog.Warnf("Config file not found. Use %s instead.", config.FallbackConfigPath)
			cfgFilePath = config.FallbackConfigPath
		}

		if _, err := os.Stat(cfgFilePath); err == nil && !overwriteFile {
			slog.Errorf("Config file %s already exists. Use --overwrite to allow overwrite this file.", cfgFilePath)
			return nil
		}

		if err := os.WriteFile(cfgFilePath, data, 0644); err != nil {
			return err
		}

		slog.Infof("Config file generated successfully at %s.", cfgFilePath)
		return nil
	},
}

func init() {
	initConfigCmd.Flags().BoolVar(&overwriteFile, "overwrite", false, "Allow overwrite existing config file")
	rootCmd.AddCommand(initConfigCmd)
}
