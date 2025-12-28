package cmd

import (
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	clh_proto "github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/internal/client/receiver"
	"github.com/sydneyowl/clh-server/pkg/msg"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	rServerIp       string
	rServerPort     int
	rUseTLS         bool
	rkey            string
	rSkipVerifyCert bool

	senderID string
)

// runReceiver represents the run-receiver command

var runReceiver = &cobra.Command{
	Use:   "run-receiver",
	Short: "Run receiver example.",
	Long:  `Run receiver example. This is for test usage only.`,
	Run: func(cmd *cobra.Command, args []string) {
		if senderID == "" {
			slog.Fatalf("sender id is required")
			return
		}
		receiver := receiver.NewReceiver(rServerIp, rServerPort, rUseTLS, rkey, rSkipVerifyCert)
		if err := receiver.DoConn(); err != nil {
			slog.Fatalf("Connection failed: %v", err)
			return
		}
		if err := receiver.DoLogin(); err != nil {
			slog.Fatalf("Login failed: %v", err)
			return
		}

		receiver.Dispatcher.RegisterHandler(&clh_proto.CommandResponse{}, receiver.OnCommandResponseReceived)
		receiver.Dispatcher.RegisterHandler(&clh_proto.WsjtxMessagePacked{}, func(message msg.Message) {
			data, err := protojson.Marshal(message)
			if err != nil {
				slog.Errorf("encode message failed: %v", err)
			} else {
				slog.Infof("Received message: %s", string(data))
			}
		})
		go receiver.Run()

		command := make([]string, 0)
		command = append(command, senderID)

		_, err := receiver.SendCommandAndWait("subscribe", command, 5)
		if err != nil {
			slog.Errorf("send command failed: %v", err)
		}

		<-receiver.Dispatcher.Done()
	},
}

func init() {
	runReceiver.Flags().StringVar(&rServerIp, "ip", "127.0.0.1", "server ip")
	runReceiver.Flags().IntVar(&rServerPort, "port", 7410, "server port")
	runReceiver.Flags().BoolVar(&rUseTLS, "tls", true, "use tls")
	runReceiver.Flags().BoolVar(&rSkipVerifyCert, "skip-cert-verify", false, "skip certificate verification")
	runReceiver.Flags().StringVar(&rkey, "key", "???", "key for conn auth usage")
	runReceiver.Flags().StringVar(&senderID, "senderId", "???", "Sender ID to subscribe")
	rootCmd.AddCommand(runReceiver)
}
