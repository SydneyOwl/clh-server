package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sydneyowl/clh-server/internal/server"
	"github.com/sydneyowl/clh-server/pkg/config"
	"github.com/sydneyowl/clh-server/pkg/log"
	"github.com/sydneyowl/clh-server/pkg/version"

	"github.com/gookit/slog"
	"github.com/spf13/cobra"
)

var (
	cfgFilePath string
	verbose     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "clh-server",
	Short:   "Server application of Cloudlog Helper.",
	Long:    `Server application of Cloudlog Helper. See github.com/SydneyOwl/cloudlog-helper for more.`,
	Version: version.Full(),

	RunE: func(cmd *cobra.Command, args []string) error {
		parsedCfg, err := ParseCfg(cfgFilePath)
		if err != nil {
			return err
		}
		_, einfo := parsedCfg.Verify()
		if len(einfo) > 0 {
			return fmt.Errorf("invalid config settings. Run check-config for detail")
		}
		if err = log.SetLogger(*(parsedCfg.Log), verbose); err != nil {
			return err
		}
		defer slog.MustClose()
		sv, err := server.NewService(parsedCfg)
		if err != nil {
			return err
		}
		sv.Run(context.Background())
		return nil
	},
	SilenceUsage: true,
}

func ParseCfg(path string) (*config.Config, error) {
	if path == "" {
		slog.Warnf("Config file not found. Use %s instead.", config.FallbackConfigPath)
		path = config.FallbackConfigPath
	}
	parseConfig, err := config.ParseConfig(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("default config file not found. You can use init-config command to generate a new one")
		}
		return nil, err
	}
	return parseConfig, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFilePath, "config", "c", "", "config file (default is .clh-server.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "use verbose output(console).")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
