package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/internal/msg"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/crypto"
	msg1 "github.com/sydneyowl/clh-server/pkg/msg"
)

type Client struct {
	ClientType string

	ServerIp       string
	ServerPort     int
	UseTLS         bool
	Key            string
	SkipCertVerify bool

	Mu         sync.Mutex
	Dispatcher *msg.Dispatcher
	Conn       net.Conn
	LastAck    atomic.Int64
}

func (c *Client) DoConn() error {
	var (
		conn net.Conn
		err  error
	)
	if !c.UseTLS {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", c.ServerIp, c.ServerPort))
	} else {
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", c.ServerIp, c.ServerPort), &tls.Config{InsecureSkipVerify: c.SkipCertVerify})
	}
	if err != nil {
		return err
	}
	c.Conn = conn
	return nil
}

func (c *Client) DoLogin() error {
	tm := time.Now().Unix()
	p, _ := strutil.RandomString(5)
	err := msg.WriteMsg(c.Conn, &msgproto.HandshakeRequest{
		Os:         "Linux",
		Ver:        "0.0",
		ClientType: c.ClientType,
		AuthKey:    crypto.CalcAuthKey(c.Key, tm),
		Timestamp:  tm,
		RunId:      "Test" + c.ClientType + p,
	})
	if err != nil {
		return err
	}
	_ = c.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	readMsg, err := msg.ReadMsg(c.Conn)
	if err != nil {
		return err
	}
	response := readMsg.(*msgproto.HandshakeResponse)
	slog.Infof("Connected to server successfully. Got acked runID %s", response.RunId)
	_ = c.Conn.SetReadDeadline(time.Time{})
	c.Dispatcher = msg.NewDispatcher(c.Conn)
	c.Dispatcher.RegisterDefaultHandler(func(msg msg1.Message) {
		slog.Warnf("WTF? Unknown message received: %T ", msg)
	})
	c.Dispatcher.RegisterHandler(&msgproto.Pong{}, c.onHBReceived)
	c.Dispatcher.RegisterHandler(&msgproto.CommonResponse{}, c.onCommonResponseReceived)
	c.LastAck.Store(time.Now().Unix())
	return nil
}

func (c *Client) doHeartbeat() {
	for {
		err := msg.WriteMsg(c.Conn, &msgproto.Ping{
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			slog.Error("HB" + err.Error())
			return
		}

		time.Sleep(time.Second * 2)
	}
}

func (c *Client) onHBReceived(mm msg1.Message) {
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
	c.LastAck.Store(time.Now().Unix())
}

func (c *Client) scanHeartbeatTimeout() {
	tt := time.NewTicker(time.Second * 2)
	defer tt.Stop()
	for {
		select {
		case <-tt.C:
			lastHB := c.LastAck.Load()
			if time.Now().Unix()-lastHB > 5 {
				slog.Errorf("HB timeout")
				_ = c.Conn.Close()
				os.Exit(0)
			}
		}
	}
}

func (c *Client) onCommonResponseReceived(mm msg1.Message) {
	a, ok := mm.(*msgproto.CommonResponse)
	if !ok {
		slog.Errorf("Cannot convert message to CommonResponse: %v", mm)
		return
	}
	slog.Info(a.Message)
}

func (c *Client) Run() {
	go c.doHeartbeat()
	go c.scanHeartbeatTimeout()

	c.Dispatcher.Run()
	<-c.Dispatcher.Done()
	slog.Infof("Receiver shutting down")
}

func (c *Client) Close() {
	_ = c.Conn.Close()
}
