package sender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSender(t *testing.T) {
	s := NewSender("127.0.0.1", 7410, true, "testkey", false)
	assert.NotNil(t, s)
	assert.Equal(t, "127.0.0.1", s.ServerIp)
	assert.Equal(t, 7410, s.ServerPort)
	assert.True(t, s.UseTLS)
	assert.Equal(t, "testkey", s.Key)
	assert.Equal(t, "sender", s.ClientType)
}
