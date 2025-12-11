package server

import (
	"context"
	"github.com/sydneyowl/clh-server/msgproto"
	"net"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/internal/msg"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type Control struct {
	ctx context.Context

	runID         string
	conn          net.Conn
	msgDispatcher *msg.Dispatcher
}

func NewControl(ctx context.Context, runID string, conn net.Conn) *Control {
	ctrl := &Control{
		ctx:   ctx,
		runID: runID,
		conn:  conn,
	}
	ctrl.msgDispatcher = msg.NewDispatcher(ctrl.conn)
	ctrl.regMsgHandlers()
	return ctrl
}

func (ctrl *Control) Run() {
	go ctrl.msgDispatcher.Run()
	<-ctrl.msgDispatcher.Done()
	slog.Debugf("message dispatcher exited")
	_ = ctrl.conn.Close()
	slog.Infof("client %s exited.", ctrl.runID)
}

func (ctrl *Control) Close() error {
	_ = ctrl.conn.Close()
	return nil
}

func (ctrl *Control) regMsgHandlers() {
	ctrl.msgDispatcher.RegisterDefaultHandler(ctrl.defaultHandler)
	ctrl.msgDispatcher.RegisterHandler(&msgproto.DigModeUpdate{}, ctrl.digModeMsgHandler)
}

func (ctrl *Control) defaultHandler(msg msg1.Message) {
	slog.Debugf("Client sent a message (%s) but no handler is registered for it...", msg)
}

func (ctrl *Control) digModeMsgHandler(msg msg1.Message) {
	slog.Tracef("Client %s sent a digmode message.", ctrl.runID)
}
