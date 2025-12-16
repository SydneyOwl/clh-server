package server

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/sydneyowl/clh-server/internal/cache"

	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/config"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/internal/msg"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type ClientType string

const (
	receiver ClientType = "receiver"
	sender   ClientType = "sender"
)

func ValidateClientType(input string) (ClientType, error) {
	ct := ClientType(input)
	switch ct {
	case receiver, sender:
		return ct, nil
	default:
		return "", fmt.Errorf("invalid ClientType: %s", input)
	}
}

type Control struct {
	ctx context.Context

	clientType    ClientType
	serverCfg     *config.Config
	cache         cache.CLHCache
	cacheToken    any
	lastPing      atomic.Value
	doneCh        chan struct{}
	runID         string
	conn          net.Conn
	msgDispatcher *msg.Dispatcher
}

func NewControl(cfg *config.Config, runID string, clientType ClientType, cache cache.CLHCache, conn net.Conn, ctx context.Context) *Control {
	ctrl := &Control{
		ctx:        ctx,
		runID:      runID,
		conn:       conn,
		cache:      cache,
		clientType: clientType,
		serverCfg:  cfg,
		doneCh:     make(chan struct{}),
	}
	ctrl.lastPing.Store(time.Now())
	ctrl.msgDispatcher = msg.NewDispatcher(ctrl.conn)
	if ctrl.clientType == sender {
		ctrl.regSenderMsgHandlers()
	}
	if ctrl.clientType == receiver {
		ctrl.regReceiverMsgHandlers()
	}
	return ctrl
}

func (ctrl *Control) Run() {
	go ctrl.heartbeatChecker()
	go ctrl.msgDispatcher.Run()
	<-ctrl.msgDispatcher.Done()
	slog.Debugf("%s message dispatcher exited", ctrl.runID)
	_ = ctrl.conn.Close()
	close(ctrl.doneCh)
	slog.Infof("client %s exited.", ctrl.runID)
}

func (ctrl *Control) Replace(newCtl *Control) {
	ctrl.runID = ctrl.runID + "-OLD"
	_ = ctrl.conn.Close()
}

func (ctrl *Control) WaitForQuit() {
	<-ctrl.doneCh
	if ctrl.clientType == receiver {
		ctrl.cache.UnsubscribeHandler(ctrl.runID, ctrl.cacheToken)
	}
	if ctrl.clientType == sender {
		ctrl.cache.RemoveCache(ctrl.runID)
	}
	slog.Tracef("Release cache of %s", ctrl.runID)
}

func (ctrl *Control) Close() error {
	_ = ctrl.conn.Close()
	return nil
}

func (ctrl *Control) regReceiverMsgHandlers() {
	slog.Tracef("Registering message handlers for receiver")
	ctrl.msgDispatcher.RegisterHandler(&msgproto.Ping{}, ctrl.heartbeatHandler)
	ctrl.msgDispatcher.RegisterHandler(&msgproto.Command{}, ctrl.commandHandler)
	ctrl.msgDispatcher.RegisterHandler(&msgproto.Subscribe{}, ctrl.subscribeHandler)
	ctrl.msgDispatcher.RegisterDefaultHandler(ctrl.defaultHandler)
}

func (ctrl *Control) regSenderMsgHandlers() {
	slog.Tracef("Registering message handlers for sender")
	ctrl.msgDispatcher.RegisterHandler(&msgproto.Ping{}, ctrl.heartbeatHandler)
	ctrl.msgDispatcher.RegisterHandler(&msgproto.WsjtxMessage{}, ctrl.wsjtxMsgHandler)
	ctrl.msgDispatcher.RegisterDefaultHandler(ctrl.defaultHandler)
}

func (ctrl *Control) defaultHandler(msg msg1.Message) {
	slog.Debugf("Client sent a message (%s) but no handler is registered for it...", msg)
}

func (ctrl *Control) heartbeatHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a heartbeat message.", ctrl.runID)
	ping, err := tryUnwrapMessage[*msgproto.Ping](msg, ctrl.msgDispatcher)
	if err {
		return
	}
	if (ping.Timestamp - time.Now().Unix()) > 10 {
		_ = ctrl.msgDispatcher.Send(&msgproto.Pong{
			Ack:   false,
			Error: "Invalid timstamp. Check ur clock!",
		})
		return
	}

	ctrl.lastPing.Store(time.Now())
	slog.Tracef("%s heartbeat message accepted.", ctrl.runID)
	_ = ctrl.msgDispatcher.Send(&msgproto.Pong{
		Ack:   true,
		Error: "",
	})
}

func (ctrl *Control) wsjtxMsgHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a wsjtx message.", ctrl.runID)
	sub, err := tryUnwrapMessage[*msgproto.WsjtxMessage](msg, ctrl.msgDispatcher)
	if err {
		return
	}
	switch v := sub.Payload.(type) {
	case *msgproto.WsjtxMessage_Status:
		_ = ctrl.cache.PublishMessage(ctrl.runID, v.Status)
	case *msgproto.WsjtxMessage_Decode:
		_ = ctrl.cache.PublishMessage(ctrl.runID, v.Decode)
	case *msgproto.WsjtxMessage_WsprDecode:
		_ = ctrl.cache.PublishMessage(ctrl.runID, v.WsprDecode)
	default:
		slog.Errorf("Unknown message type: %T", sub.Payload)
	}
}

func (ctrl *Control) subscribeHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a subscribe message.", ctrl.runID)
	sub, err := tryUnwrapMessage[*msgproto.Subscribe](msg, ctrl.msgDispatcher)
	if err {
		return
	}
	if ctrl.cacheToken != nil {
		ctrl.cache.UnsubscribeHandler(ctrl.runID, ctrl.cacheToken)
	}
	ctrl.cacheToken = ctrl.cache.SubscribeHandler(sub.TargetRunId, ctrl.doForward)
}

func (ctrl *Control) commandHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a command message.", ctrl.runID)
	//cmd, err := tryUnwrapMessage[*msgproto.Command](msg, ctrl.msgDispatcher)
	//if err {return}
}

func (ctrl *Control) heartbeatChecker() {
	if ctrl.serverCfg.Server.Transport.HeartbeatTimeoutSec <= 0 {
		return
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctrl.doneCh:
			slog.Tracef("Stop checking heartbeat for %s.", ctrl.runID)
			return
		case <-ticker.C:
			// check heartbeat
			if time.Since(ctrl.lastPing.Load().(time.Time)).Seconds() > ctrl.serverCfg.Server.Transport.HeartbeatTimeoutSec {
				slog.Warnf("Heartbeat timeout for %s.", ctrl.runID)
				_ = ctrl.Close()
				return
			}
		}
	}
}

func (ctrl *Control) doForward(message msg1.Message) {
	panic("Impl1!!")
}

func tryUnwrapMessage[T any](msg msg1.Message, dispatcher *msg.Dispatcher) (T, bool) {
	var zero T
	unwrapped, ok := msg.(T)
	if !ok {
		_ = dispatcher.Send(&msgproto.CommonResponse{
			Success: false,
			Message: "Invalid message type",
		})
		return zero, false
	}
	return unwrapped, true
}
