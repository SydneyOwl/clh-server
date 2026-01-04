package server

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clh_proto "github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/internal/msg"
	"github.com/sydneyowl/clh-server/pkg/config"
	"github.com/sydneyowl/clh-server/pkg/crypto"
)

func TestServerClientHandshake(t *testing.T) {
	// Create a config for test
	cfg := config.GenerateDefaultConfig()
	cfg.Server.BindAddr = "127.0.0.1"
	cfg.Server.BindPort = 0              // Use 0 for auto assign
	cfg.Server.Encrypt.EnableTLS = false // Disable TLS for simplicity
	cfg.Server.Encrypt.Key = "testkey12345"

	// Create server
	svr, err := NewService(cfg)
	assert.NoError(t, err)

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	go svr.Run(ctx)

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual port
	addr := svr.listener.Addr().(*net.TCPAddr)
	port := addr.Port

	// Connect client
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	assert.NoError(t, err)
	defer conn.Close()

	// Send handshake request
	tm := time.Now().Unix()
	authKey := crypto.CalcAuthKey(cfg.Server.Encrypt.Key, tm)
	req := &clh_proto.HandshakeRequest{
		Os:         "Linux",
		Ver:        "0.0.1",
		ClientType: "receiver",
		AuthKey:    authKey,
		Timestamp:  tm,
		RunId:      "testrun",
	}
	err = msg.WriteMsg(conn, req)
	assert.NoError(t, err)

	// Read response
	respMsg, err := msg.ReadMsg(conn)
	assert.NoError(t, err)
	resp, ok := respMsg.(*clh_proto.HandshakeResponse)
	assert.True(t, ok)
	assert.True(t, resp.Accept)
	assert.Equal(t, "testrun", resp.RunId)

	// Stop server
	cancel()
	time.Sleep(100 * time.Millisecond)
}
