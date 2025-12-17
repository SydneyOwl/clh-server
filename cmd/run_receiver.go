package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sydneyowl/clh-server/internal/client/receiver"
)

var (
	serverIp       string
	serverPort     int
	useTLS         bool
	key            string
	skipVerifyCert bool
)

// checkConfigCmd represents the checkConfig command
var runReceiver = &cobra.Command{
	Use:   "run-receiver",
	Short: "Run receiver example.",
	Long:  `Run receiver example. This is for test usage only.`,
	Run: func(cmd *cobra.Command, args []string) {
		receiver := receiver.NewReceiver(serverIp, serverPort, useTLS, key, skipVerifyCert)
		receiver.Run()
	},
}

func init() {
	runReceiver.Flags().StringVar(&serverIp, "ip", "127.0.0.1", "server ip")
	runReceiver.Flags().IntVar(&serverPort, "port", 7410, "server port")
	runReceiver.Flags().BoolVar(&useTLS, "tls", true, "use tls")
	runReceiver.Flags().BoolVar(&skipVerifyCert, "skip-cert-verify", false, "skip certificate verification")
	runReceiver.Flags().StringVar(&key, "key", "???", "key for conn auth usage")
	rootCmd.AddCommand(runReceiver)
}
