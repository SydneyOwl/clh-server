package server

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/config"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/internal/msg"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type Control struct {
	ctx context.Context

	serverCfg     *config.Config
	lastPing      atomic.Value
	doneCh        chan struct{}
	runID         string
	conn          net.Conn
	msgDispatcher *msg.Dispatcher
}

func NewControl(cfg *config.Config, runID string, conn net.Conn, ctx context.Context) *Control {
	ctrl := &Control{
		ctx:       ctx,
		runID:     runID,
		conn:      conn,
		serverCfg: cfg,
		doneCh:    make(chan struct{}),
	}
	ctrl.lastPing.Store(time.Now())
	ctrl.msgDispatcher = msg.NewDispatcher(ctrl.conn)
	ctrl.regMsgHandlers()
	return ctrl
}

func (ctrl *Control) Run() {
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
}

func (ctrl *Control) Close() error {
	_ = ctrl.conn.Close()
	return nil
}

func (ctrl *Control) regMsgHandlers() {
	go ctrl.heartbeatChecker()
	// todo waitgroup?
	ctrl.msgDispatcher.RegisterDefaultHandler(ctrl.defaultHandler)
	ctrl.msgDispatcher.RegisterHandler(&msgproto.WsjtxMessage{}, ctrl.wsjtxMsgHandler)
	ctrl.msgDispatcher.RegisterHandler(&msgproto.Ping{}, ctrl.heartbeatHandler)
}

func (ctrl *Control) defaultHandler(msg msg1.Message) {
	slog.Debugf("Client sent a message (%s) but no handler is registered for it...", msg)
}

func (ctrl *Control) heartbeatHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a heartbeat message.", ctrl.runID)
	ping, ok := msg.(*msgproto.Ping)
	if !ok {
		_ = ctrl.msgDispatcher.Send(&msgproto.Pong{
			Ack:   false,
			Error: "Invalid message type",
		})
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
