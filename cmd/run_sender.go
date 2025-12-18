package cmd

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/clh-server/internal/client/pkg"
	"github.com/sydneyowl/clh-server/internal/client/sender"
)

var (
	sServerIp       string
	sServerPort     int
	sUseTLS         bool
	skey            string
	sSkipVerifyCert bool

	doneChan chan struct{}
)

// checkConfigCmd represents the checkConfig command
var runSender = &cobra.Command{
	Use:   "run-sender",
	Short: "Run sender example.",
	Long:  `Run sender example. This is for test usage only.`,
	Run: func(cmd *cobra.Command, args []string) {
		sender := sender.NewSender(sServerIp, sServerPort, sUseTLS, skey, sSkipVerifyCert)
		if err := sender.DoConn(); err != nil {
			slog.Fatalf("Conn failed: %v", err)
			return
		}
		if err := sender.DoLogin(); err != nil {
			slog.Fatalf("Login failed: %v", err)
			return
		}
		p := tea.NewProgram(pkg.InitialModel(sender.Run, func(command string) (string, error) {
			return "", nil
		}, func() {
			doneChan <- struct{}{}
			sender.Close()
		}))

		go sender.TestRandomInfoSending(doneChan)

		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	doneChan = make(chan struct{})

	runSender.Flags().StringVar(&sServerIp, "ip", "127.0.0.1", "server ip")
	runSender.Flags().IntVar(&sServerPort, "port", 7410, "server port")
	runSender.Flags().BoolVar(&sUseTLS, "tls", true, "use tls")
	runSender.Flags().BoolVar(&sSkipVerifyCert, "skip-cert-verify", false, "skip certificate verification")
	runSender.Flags().StringVar(&skey, "key", "???", "key for conn auth usage")
	rootCmd.AddCommand(runSender)
}
