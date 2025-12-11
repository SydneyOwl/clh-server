package cmd

import (
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
)

// checkConfigCmd represents the checkConfig command
var checkConfigCmd = &cobra.Command{
	Use:   "check-config",
	Short: "Check if specified config file is valid.",
	Long:  `check if specified config file is valid.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := ParseCfg(cfgFilePath)
		if err != nil {
			slog.Fatalf("Error parsing config file: %s", err)
		}
		warnMsg, errMsg := cfg.Verify()
		if len(warnMsg) == 0 && len(errMsg) == 0 {
			slog.Info("Config file OK.")
			return
		}
		slog.Notice("-------RESULT------")
		for _, s := range warnMsg {
			slog.Warn(s)
		}
		for _, v := range errMsg {
			slog.Error(v)
		}
		slog.Notice("-------------------")
	},
}

func init() {
	rootCmd.AddCommand(checkConfigCmd)
}
