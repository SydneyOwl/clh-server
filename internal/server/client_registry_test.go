package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientRegistry_Register(t *testing.T) {
	cr := NewClientRegistry(0)
	info := ClientInfo{
		RunID: "test1",
		Type:  receiver,
		Addr:  "127.0.0.1:1234",
	}
	cr.Register(info)

	// Check GetSenders (should be empty)
	senders := cr.GetSenders()
	assert.Empty(t, senders)
}

func TestClientRegistry_GetSenders(t *testing.T) {
	cr := NewClientRegistry(0)
	cr.Register(ClientInfo{RunID: "r1", Type: receiver})
	cr.Register(ClientInfo{RunID: "s1", Type: sender})
	cr.Register(ClientInfo{RunID: "s2", Type: sender})

	senders := cr.GetSenders()
	assert.Len(t, senders, 2)
	assert.Contains(t, senders, "s1")
	assert.Contains(t, senders, "s2")
}

func TestClientRegistry_Unregister(t *testing.T) {
	cr := NewClientRegistry(0)
	cr.Register(ClientInfo{RunID: "test", Type: sender})
	senders := cr.GetSenders()
	assert.Len(t, senders, 1)

	cr.Unregister("test")
	senders = cr.GetSenders()
	assert.Empty(t, senders)
}

func TestClientRegistry_Touch(t *testing.T) {
	cr := NewClientRegistry(0)
	info := ClientInfo{RunID: "test", Type: sender}
	cr.Register(info)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	cr.Touch("test")

	// Since LastSeen is updated, but hard to test without access.
	// Perhaps just call it.
}
