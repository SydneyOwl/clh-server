package receiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReceiver(t *testing.T) {
	r := NewReceiver("127.0.0.1", 7410, true, "testkey", false)
	assert.NotNil(t, r)
	assert.Equal(t, "127.0.0.1", r.ServerIp)
	assert.Equal(t, 7410, r.ServerPort)
	assert.True(t, r.UseTLS)
	assert.Equal(t, "testkey", r.Key)
	assert.Equal(t, "receiver", r.ClientType)
	assert.NotNil(t, r.commandPendingMap)
}
