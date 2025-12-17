package receiver

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/internal/msg"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/crypto"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type Receiver struct {
	serverIp       string
	serverPort     int
	useTLS         bool
	key            string
	skipCertVerify bool

	commandPendingMap map[string]chan<- string

	mu         sync.Mutex
	dispatcher *msg.Dispatcher
	conn       net.Conn
}

func NewReceiver(serverIp string, serverPort int, useTLS bool, key string, skipCertVerify bool) *Receiver {
	return &Receiver{
		serverIp:          serverIp,
		serverPort:        serverPort,
		useTLS:            useTLS,
		key:               key,
		commandPendingMap: make(map[string]chan<- string),
		skipCertVerify:    skipCertVerify,
	}
}

func (r *Receiver) doConn() error {
	var (
		conn net.Conn
		err  error
	)
	if !r.useTLS {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", r.serverIp, r.serverPort))
	} else {
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", r.serverIp, r.serverPort), &tls.Config{InsecureSkipVerify: r.skipCertVerify})
	}
	if err != nil {
		return err
	}
	r.conn = conn
	return nil
}

func (r *Receiver) doLogin() error {
	tm := time.Now().Unix()
	p, _ := strutil.RandomString(5)
	err := msg.WriteMsg(r.conn, &msgproto.HandshakeRequest{
		Os:         "Linux",
		Ver:        "0.0",
		ClientType: "receiver",
		AuthKey:    crypto.CalcAuthKey(r.key, tm),
		Timestamp:  tm,
		RunId:      "TestReceiver - " + p,
	})
	if err != nil {
		return err
	}
	_ = r.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	readMsg, err := msg.ReadMsg(r.conn)
	if err != nil {
		return err
	}
	response := readMsg.(*msgproto.HandshakeResponse)
	slog.Infof("Connected to server successfully. Got acked runID %s", response.RunId)
	_ = r.conn.SetReadDeadline(time.Time{})
	r.dispatcher = msg.NewDispatcher(r.conn)
	r.dispatcher.RegisterDefaultHandler(func(msg msg1.Message) {
		slog.Warnf("WTF? Unknown message received: %T ", msg)
	})
	r.dispatcher.RegisterHandler(&msgproto.Pong{}, r.onHBReceived)
	r.dispatcher.RegisterHandler(&msgproto.CommonResponse{}, r.onCommonResponseReceived)
	r.dispatcher.RegisterHandler(&msgproto.CommandResponse{}, r.onCommandResponseReceived)
	r.dispatcher.Run()
	return nil
}

func (r *Receiver) doHeartbeat() {
	for {
		err := msg.WriteMsg(r.conn, &msgproto.Ping{
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			slog.Error("HB" + err.Error())
			return
		}

		time.Sleep(time.Second * 5)
	}
}

func (r *Receiver) onHBReceived(mm msg1.Message) {
	a, ok := mm.(*msgproto.Pong)
	if !ok {
		slog.Errorf("Cannot convert message to Pong: %v", mm)
		return
	}
	if !a.Ack {
		slog.Errorf("Server declined out heartbeat: %s", a.Error)
		return
	}
	slog.Tracef("Server accepted our heartbeat.")
}

func (r *Receiver) onCommonResponseReceived(mm msg1.Message) {
	a, ok := mm.(*msgproto.CommonResponse)
	if !ok {
		slog.Errorf("Cannot convert message to CommonResponse: %v", mm)
		return
	}
	slog.Info(a.Message)
}

func (r *Receiver) onCommandResponseReceived(mm msg1.Message) {
	a, ok := mm.(*msgproto.CommandResponse)
	if !ok {
		slog.Errorf("Cannot convert message to CommandResponse: %v", mm)
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.mu.Lock()
	r.commandPendingMap[ranId] = cc
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		close(cc)
		delete(r.commandPendingMap, ranId)
	}()

	err := r.dispatcher.Send(&msgproto.Command{
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

func (r *Receiver) Run() {
	if err := r.doConn(); err != nil {
		slog.Fatalf("Conn failed: %v", err)
		return
	}
	if err := r.doLogin(); err != nil {
		slog.Fatalf("Login failed: %v", err)
		return
	}
	go r.doHeartbeat()

	<-r.dispatcher.Done()
	slog.Infof("Receiver shutting down")
}

func (r *Receiver) Close() {
	_ = r.conn.Close()
}
