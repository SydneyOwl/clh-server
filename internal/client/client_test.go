package client

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_DoConn(t *testing.T) {
	// Start a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	// Start accepting in background
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	client := &Client{
		ServerIp:   "127.0.0.1",
		ServerPort: port,
		UseTLS:     false,
	}

	err = client.DoConn()
	assert.NoError(t, err)
	assert.NotNil(t, client.Conn)
	client.Conn.Close()
}
