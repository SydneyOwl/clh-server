package cmd

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gookit/slog"
	"github.com/spf13/cobra"
	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/internal/client/pkg"
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
)

// runReceiver represents the run-receiver command
var runReceiver = &cobra.Command{
	Use:   "run-receiver",
	Short: "Run receiver example.",
	Long:  `Run receiver example. This is for test usage only.`,
	Run: func(cmd *cobra.Command, args []string) {
		receiver := receiver.NewReceiver(rServerIp, rServerPort, rUseTLS, rkey, rSkipVerifyCert)
		if err := receiver.DoConn(); err != nil {
			slog.Fatalf("Connection failed: %v", err)
			return
		}
		if err := receiver.DoLogin(); err != nil {
			slog.Fatalf("Login failed: %v", err)
			return
		}

		model := pkg.InitialModel(receiver.Run, func(command string) (string, error) {
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
		}, receiver.Close)

		var program *tea.Program
		program = tea.NewProgram(model) // fixme lock problem

		receiver.Dispatcher.RegisterHandler(&clh_proto.CommandResponse{}, receiver.OnCommandResponseReceived)
		receiver.Dispatcher.RegisterHandler(&clh_proto.WsjtxMessagePacked{}, func(message msg.Message) {
			data, err := protojson.Marshal(message)
			if err != nil {
				program.Send(pkg.ResponseMsg{Response: fmt.Sprintf("Error marshaling message: %v", err), Err: nil})
			} else {
				program.Send(pkg.ResponseMsg{Response: string(data), Err: nil})
			}
		})
		if _, err := program.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	runReceiver.Flags().StringVar(&rServerIp, "ip", "127.0.0.1", "server ip")
	runReceiver.Flags().IntVar(&rServerPort, "port", 7410, "server port")
	runReceiver.Flags().BoolVar(&rUseTLS, "tls", true, "use tls")
	runReceiver.Flags().BoolVar(&rSkipVerifyCert, "skip-cert-verify", false, "skip certificate verification")
	runReceiver.Flags().StringVar(&rkey, "key", "???", "key for conn auth usage")
	rootCmd.AddCommand(runReceiver)
}
