package server

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sydneyowl/clh-server/internal/cache"
	"github.com/sydneyowl/clh-server/pkg/debounce"

	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/pkg/config"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/internal/msg"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type ClientType string
type CommandType string

const (
	debounceInterval = time.Second * 2

	receiver ClientType = "receiver"
	sender   ClientType = "sender"

	subscribe     CommandType = "subscribe"
	unsubscribe   CommandType = "unsubscribe"
	getSenderList CommandType = "getSenderList"
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

	clientType     ClientType
	serverCfg      *config.Config
	cache          cache.CLHCache
	clientRegistry *ClientRegistry

	subscribedId string
	cacheToken   any

	lastPing      atomic.Value
	doneCh        chan struct{}
	runID         string
	conn          net.Conn
	closeOnce     *sync.Once
	msgDispatcher *msg.Dispatcher

	msgDebouncer *debounce.Debouncer

	msgQueueLocker sync.Mutex
	msgSendQueue   []msg1.Message
}

func NewControl(cfg *config.Config, runID string, clientType ClientType,
	cache cache.CLHCache, conn net.Conn, ctx context.Context, registry *ClientRegistry) *Control {
	ctrl := &Control{
		ctx:            ctx,
		runID:          runID,
		conn:           conn,
		cache:          cache,
		clientType:     clientType,
		serverCfg:      cfg,
		closeOnce:      &sync.Once{},
		doneCh:         make(chan struct{}),
		msgDebouncer:   debounce.NewDebouncer(debounceInterval),
		msgSendQueue:   make([]msg1.Message, 0),
		clientRegistry: registry,
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

func (ctrl *Control) GetClientType() ClientType {
	return ctrl.clientType
}

func (ctrl *Control) Run() {
	go ctrl.heartbeatChecker()
	go ctrl.msgDispatcher.Run()
	<-ctrl.msgDispatcher.Done()
	slog.Debugf("%s message dispatcher exited", ctrl.runID)
	_ = ctrl.Close()
	slog.Infof("client %s exited.", ctrl.runID)
}

func (ctrl *Control) Replace(newCtl *Control) {
	ctrl.runID = ctrl.runID + "-OLD"
	_ = ctrl.conn.Close()
}

func (ctrl *Control) WaitForQuit() {
	<-ctrl.doneCh
	if ctrl.clientType == receiver && ctrl.cacheToken != nil {
		ctrl.cache.UnsubscribeHandler(ctrl.subscribedId, ctrl.cacheToken)
	}
	if ctrl.clientType == sender {
		ctrl.cache.RemoveCache(ctrl.runID)
	}
	slog.Tracef("Release cache of %s", ctrl.runID)
}

func (ctrl *Control) Close() error {
	ctrl.closeOnce.Do(func() {
		close(ctrl.doneCh)
		_ = ctrl.conn.Close()
	})
	return nil
}

func (ctrl *Control) regReceiverMsgHandlers() {
	slog.Tracef("Registering message handlers for receiver")
	ctrl.msgDispatcher.RegisterHandler(&clh_proto.Ping{}, ctrl.heartbeatHandler)
	ctrl.msgDispatcher.RegisterHandler(&clh_proto.Command{}, ctrl.commandHandler)
	ctrl.msgDispatcher.RegisterDefaultHandler(ctrl.defaultHandler)
}

func (ctrl *Control) regSenderMsgHandlers() {
	slog.Tracef("Registering message handlers for sender")
	ctrl.msgDispatcher.RegisterHandler(&clh_proto.Ping{}, ctrl.heartbeatHandler)
	ctrl.msgDispatcher.RegisterHandler(&clh_proto.WsjtxMessage{}, ctrl.wsjtxMsgHandler)
	ctrl.msgDispatcher.RegisterHandler(&clh_proto.RigData{}, ctrl.rigDataMsgHander)
	ctrl.msgDispatcher.RegisterDefaultHandler(ctrl.defaultHandler)
}

// handlers

func (ctrl *Control) defaultHandler(msg msg1.Message) {
	slog.Debugf("Client sent a message (%s) but no handler is registered for it...", msg)
}

func (ctrl *Control) heartbeatHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a heartbeat message.", ctrl.runID)
	ping, success := tryUnwrapMessage[*clh_proto.Ping](msg, ctrl.msgDispatcher)
	if !success {
		return
	}
	if math.Abs(float64(ping.Timestamp-time.Now().Unix())) > 10 {
		_ = ctrl.msgDispatcher.Send(&clh_proto.Pong{
			Ack:   false,
			Error: "Invalid timestamp. Check your clock!",
		})
		return
	}

	ctrl.lastPing.Store(time.Now())
	slog.Tracef("%s heartbeat message accepted.", ctrl.runID)
	_ = ctrl.msgDispatcher.Send(&clh_proto.Pong{
		Ack:   true,
		Error: "",
	})
}

func (ctrl *Control) wsjtxMsgHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a wsjtx message.", ctrl.runID)
	sub, success := tryUnwrapMessage[*clh_proto.WsjtxMessage](msg, ctrl.msgDispatcher)
	if !success {
		return
	}
	var prodErr error
	switch v := sub.Payload.(type) {
	case *clh_proto.WsjtxMessage_Status:
		prodErr = ctrl.cache.PublishMessage(ctrl.runID, v.Status)
	case *clh_proto.WsjtxMessage_Decode:
		prodErr = ctrl.cache.PublishMessage(ctrl.runID, v.Decode)
	case *clh_proto.WsjtxMessage_WsprDecode:
		prodErr = ctrl.cache.PublishMessage(ctrl.runID, v.WsprDecode)
	default:
		prodErr = fmt.Errorf("unknown message type %T", sub.Payload)
	}
	if prodErr != nil {
		slog.Error(prodErr.Error())
		//_ = ctrl.respondCommand(false, prodErr.Error())
		return
	}
	//_ = ctrl.respondCommand(true, "")
}

func (ctrl *Control) rigDataMsgHander(msg msg1.Message) {
	sub, success := tryUnwrapMessage[*clh_proto.RigData](msg, ctrl.msgDispatcher)
	if !success {
		return
	}
	if prodErr := ctrl.cache.PublishMessage(ctrl.runID, sub); prodErr != nil {
		slog.Error(prodErr.Error())
		return
	}
}

func (ctrl *Control) commandHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a command message.", ctrl.runID)
	command, success := tryUnwrapMessage[*clh_proto.Command](msg, ctrl.msgDispatcher)
	if !success {
		return
	}
	switch CommandType(command.Command) {
	case subscribe:
		if len(command.Params) < 1 {
			_ = ctrl.respondCommand(false, "Invalid command parameters.", command.CommandId)
			return
		}
		ctrl.subscribe(command.Params[0])
		_ = ctrl.respondCommand(true, "", command.CommandId)
	case unsubscribe:
		ctrl.unsubscribe()
		_ = ctrl.respondCommand(true, "", command.CommandId)
	case getSenderList:
		ls := ctrl.clientRegistry.GetSenders()
		// todo for now we use this separator...
		res := strings.Join(ls, ",,,")
		_ = ctrl.respondCommand(true, res, command.CommandId)
	}
}

// shared

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
			if t, ok := ctrl.lastPing.Load().(time.Time); ok {
				if time.Since(t).Seconds() > ctrl.serverCfg.Server.Transport.HeartbeatTimeoutSec {
					slog.Warnf("Heartbeat timeout for %s.", ctrl.runID)
					_ = ctrl.Close()
					return
				}
			} else {
				slog.Errorf("This should never happen...")
			}
		}
	}
}

func (ctrl *Control) subscribe(targetRunId string) {
	ctrl.unsubscribe()
	ctrl.subscribedId = targetRunId
	ctrl.cacheToken = ctrl.cache.SubscribeHandler(targetRunId, ctrl.doForward)
}

func (ctrl *Control) unsubscribe() {
	if ctrl.cacheToken != nil {
		ctrl.cache.UnsubscribeHandler(ctrl.subscribedId, ctrl.cacheToken)
		ctrl.cacheToken = nil
		ctrl.subscribedId = ""
	}
}

func (ctrl *Control) doForward(message []msg1.Message) {
	ctrl.msgQueueLocker.Lock()
	ctrl.msgSendQueue = append(ctrl.msgSendQueue, message...)
	ctrl.msgQueueLocker.Unlock()

	ctrl.msgDebouncer.Call(func(i any) {
		ctrl.msgQueueLocker.Lock()
		defer ctrl.msgQueueLocker.Unlock()
		slog.Tracef("Forwarding message to %s.", ctrl.runID)
		res, ok := i.(*[]msg1.Message)
		if !ok {
			slog.Warnf("Unexpected array detected while debouncing!")
			return
		}

		qlen := len(*res)
		if qlen == 0 {
			return
		}

		tbs := &clh_proto.WsjtxMessagePacked{
			Status:      nil,
			Decodes:     nil,
			WsprDecodes: nil,
		}
		// pack msgs
		for _, val := range *res {
			switch vv := val.(type) {
			case *clh_proto.Status:
				tbs.Status = vv
			case *clh_proto.Decode:
				if tbs.Decodes == nil {
					tbs.Decodes = make([]*clh_proto.Decode, 0, qlen)
				}
				tbs.Decodes = append(tbs.Decodes, vv)
			case *clh_proto.WSPRDecode:
				if tbs.WsprDecodes == nil {
					tbs.WsprDecodes = make([]*clh_proto.WSPRDecode, 0, qlen)
				}
				tbs.WsprDecodes = append(tbs.WsprDecodes, vv)
			}
		}
		// remove queue
		*res = (*res)[:0]

		if err := ctrl.msgDispatcher.Send(tbs); err != nil {
			slog.Warnf("Failed to forward message to %s: %s", ctrl.runID, err)
		}

		slog.Tracef("%v", tbs)
	}, &ctrl.msgSendQueue)
}

func (ctrl *Control) respondCommand(success bool, result string, commandId string) error {
	return ctrl.msgDispatcher.Send(&clh_proto.CommandResponse{
		CommandId: commandId,
		Success:   success,
		Result:    result,
	})
}

func tryUnwrapMessage[T any](msg msg1.Message, dispatcher *msg.Dispatcher) (T, bool) {
	var zero T
	unwrapped, ok := msg.(T)
	if !ok {
		_ = dispatcher.Send(&clh_proto.CommonResponse{
			Success: false,
			Message: "Invalid message type",
		})
		return zero, false
	}
	return unwrapped, true
}
