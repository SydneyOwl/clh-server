package cmd

import (
	"github.com/gookit/slog"
	"github.com/k0swe/wsjtx-go/v4"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var (
	sServerIp       string
	sServerPort     int
	sUseTLS         bool
	skey            string
	sSkipVerifyCert bool

	wsjtxUdp  bool = true
	randomMsg bool
)

// runSender represents the run-sender command
var runSender = &cobra.Command{
	Use:   "run-sender",
	Short: "Run sender example.",
	Long:  `Run sender example. This is for test usage only.`,
	Run: func(cmd *cobra.Command, args []string) {
		//sender := sender.NewSender(sServerIp, sServerPort, sUseTLS, skey, sSkipVerifyCert)
		//if err := sender.DoConn(); err != nil {
		//	slog.Fatalf("Connection failed: %v", err)
		//	return
		//}
		//if err := sender.DoLogin(); err != nil {
		//	slog.Fatalf("Login failed: %v", err)
		//	return
		//}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		if randomMsg {
			//go sender.TestRandomInfoSending(sigChan)
		} else {
			wsjtxServer, err := wsjtx.MakeServer()
			if err != nil {
				slog.Fatalf("%v", err)
				return
			}
			slog.Infof("Wsjtx server created successfully at 127.0.0.1:2237")
			wsjtxChannel := make(chan interface{}, 5)
			errChannel := make(chan error, 5)
			go wsjtxServer.ListenToWsjtx(wsjtxChannel, errChannel)
			go func(msgChan chan interface{}, errChan chan error) {
				for {
					select {
					case msg := <-msgChan:
						slog.Infof("Got msg: %v", msg)
					case err := <-errChannel:
						slog.Errorf("%v", err)
					}
				}
			}(wsjtxChannel, errChannel)
		}

		<-sigChan
	},
}

func init() {
	runSender.Flags().StringVar(&sServerIp, "ip", "127.0.0.1", "server ip")
	runSender.Flags().IntVar(&sServerPort, "port", 7410, "server port")
	runSender.Flags().BoolVar(&sUseTLS, "tls", true, "use tls")
	runSender.Flags().BoolVar(&sSkipVerifyCert, "skip-cert-verify", false, "skip certificate verification")
	runSender.Flags().StringVar(&skey, "key", "???", "key for conn auth usage")

	runSender.Flags().BoolVar(&wsjtxUdp, "wsjtx-udp", true, "upload messages received from wsjtx client.")
	runSender.Flags().BoolVar(&randomMsg, "random-msg", false, "upload messages generated randomly.")

	runSender.MarkFlagsOneRequired("wsjtx-udp", "random-msg")       // 至少需要一个
	runSender.MarkFlagsMutuallyExclusive("wsjtx-udp", "random-msg") // 不能同时使用
	rootCmd.AddCommand(runSender)
}
