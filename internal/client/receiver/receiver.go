package receiver

import (
	"fmt"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/internal/client"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type Receiver struct {
	client.Client

	commandPendingMap map[string]chan<- string
}

func NewReceiver(serverIp string, serverPort int, useTLS bool, key string, skipCertVerify bool) *Receiver {
	return &Receiver{
		Client: client.Client{
			ServerIp:       serverIp,
			ServerPort:     serverPort,
			UseTLS:         useTLS,
			SkipCertVerify: skipCertVerify,
			Key:            key,
			ClientType:     "receiver",
		},
		commandPendingMap: make(map[string]chan<- string),
	}
}

func (r *Receiver) OnCommandResponseReceived(mm msg1.Message) {
	a, ok := mm.(*clh_proto.CommandResponse)
	if !ok {
		slog.Errorf("Cannot convert message to CommandResponse: %v", mm)
		return
	}
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()
	res, ok := r.commandPendingMap[a.CommandId]
	if !ok {
		slog.Warnf("command response received but with an unknown command id: %v", a.CommandId)
		return
	}
	res <- a.Result
}

func (r *Receiver) SendCommandAndWait(command string, params []string, timeoutSec int) (*string, error) {
	ranId, _ := strutil.RandomString(6)
	cc := make(chan string, 1)
	r.Client.Mu.Lock()
	r.commandPendingMap[ranId] = cc
	r.Client.Mu.Unlock()

	defer func() {
		r.Client.Mu.Lock()
		defer r.Client.Mu.Unlock()
		close(cc)
		delete(r.commandPendingMap, ranId)
	}()

	err := r.Client.Dispatcher.Send(&clh_proto.Command{
		CommandId: ranId,
		Command:   command,
		Params:    params,
	})
	if err != nil {
		return nil, err
	}

	ticker := time.After(time.Second * time.Duration(timeoutSec))
	select {
	case <-ticker:
		slog.Warnf("Command %s (%s) timed out", command, ranId)
		return nil, fmt.Errorf("command %s (%s) timed out", command, ranId)
	case res, ok := <-cc:
		if !ok {
			return nil, fmt.Errorf("command channel closed")
		}
		return &res, nil
	}
}
