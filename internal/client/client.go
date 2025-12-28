package client

import (
	"crypto/tls"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gookit/goutil/strutil"
	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/internal/msg"
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
	addr := net.JoinHostPort(c.ServerIp, strconv.Itoa(c.ServerPort))
	if !c.UseTLS {
		conn, err = net.Dial("tcp", addr)
	} else {
		conn, err = tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: c.SkipCertVerify})
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
	err := msg.WriteMsg(c.Conn, &clh_proto.HandshakeRequest{
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
	response := readMsg.(*clh_proto.HandshakeResponse)
	slog.Infof("Connected to server successfully. Got acked runID %s", response.RunId)
	_ = c.Conn.SetReadDeadline(time.Time{})
	c.Dispatcher = msg.NewDispatcher(c.Conn)
	c.Dispatcher.RegisterHandler(&clh_proto.Pong{}, c.onHBReceived)
	c.Dispatcher.RegisterDefaultHandler(c.onDefaultFallbackReceived)
	c.LastAck.Store(time.Now().Unix())
	return nil
}

func (c *Client) doHeartbeat() {
	for {
		err := msg.WriteMsg(c.Conn, &clh_proto.Ping{
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
	a, ok := mm.(*clh_proto.Pong)
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
	for range tt.C {
		lastHB := c.LastAck.Load()
		if time.Now().Unix()-lastHB > 10 {
			slog.Errorf("HB timeout")
			_ = c.Conn.Close()
			os.Exit(0)
		}
	}
}

func (c *Client) onDefaultFallbackReceived(mm msg1.Message) {
	slog.Info(mm)
}

func (c *Client) Run() {
	go c.doHeartbeat()
	go c.scanHeartbeatTimeout()

	c.Dispatcher.Run()
	<-c.Dispatcher.Done()
	slog.Infof("Shutting down")
}

func (c *Client) Close() {
	_ = c.Conn.Close()
}
