package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/clh-server/internal/client/pkg"
	"github.com/sydneyowl/clh-server/internal/client/receiver"
	"log"
	"strings"
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
		p := tea.NewProgram(pkg.InitialModel(receiver.Run, func(command string) (string, error) {
			sp := strings.Split(command, " ")
			if len(sp) == 0 {
				return "", nil
			}
			if len(sp) == 1 {
				res, err := receiver.SendCommandAndWait(sp[0], nil, 1)
				if err != nil {
					return "", err
				}
				if res != nil {
					return *res, nil
				}
				return "", nil
			}
			if len(sp) >= 2 {
				res, err := receiver.SendCommandAndWait(sp[0], sp[1:], 1)
				if err != nil {
					return "", err
				}
				if res != nil {
					return *res, nil
				}
				return "", nil
			}
			return "", nil
		}, receiver.Close))

		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
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
