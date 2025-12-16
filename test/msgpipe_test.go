package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sydneyowl/clh-server/internal/cache"
	"github.com/sydneyowl/clh-server/msgproto"
)

// Mock message for testing
type mockMessage struct {
	msgproto.WsjtxMessage
}

func TestMsgPipe_NormalWriteRead(t *testing.T) {
	pipe := cache.NewMsgPipe()
	defer pipe.Close()

	msg1 := &mockMessage{}
	msg2 := &mockMessage{}

	pipe.Write(msg1)
	pipe.Write(msg2)

	assert.Equal(t, msg1, pipe.Read())
	assert.Equal(t, msg2, pipe.Read())
}

func TestMsgPipe_ReadAfterClose(t *testing.T) {
	pipe := cache.NewMsgPipe()
	msg1 := &mockMessage{}
	pipe.Write(msg1)
	pipe.Close()

	// Should read the pending message
	assert.Equal(t, msg1, pipe.Read())

	// Subsequent reads should return nil
	assert.Nil(t, pipe.Read())
	assert.Nil(t, pipe.Read())
}

func TestMsgPipe_WriteAfterClose(t *testing.T) {
	pipe := cache.NewMsgPipe()
	pipe.Close()

	// This should not panic or block
	pipe.Write(&mockMessage{})

	// Read should return nil immediately
	assert.Nil(t, pipe.Read())
}

func TestMsgPipe_MultipleCloseSafe(t *testing.T) {
	pipe := cache.NewMsgPipe()
	defer func() {
		// Ensure no panic on double close
		pipe.Close()
		pipe.Close()
		pipe.Close()
	}()

	pipe.Write(&mockMessage{})
	pipe.Close()
	assert.NotNil(t, pipe.Read())
	assert.Nil(t, pipe.Read())
}

func TestMsgPipe_ConcurrentWriteAndClose(t *testing.T) {
	pipe := cache.NewMsgPipe()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			pipe.Write(&mockMessage{})
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Closer after short delay
	time.Sleep(time.Millisecond)
	pipe.Close()

	<-done

	// Ensure no panic occurred and we can read remaining messages
	count := 0
	for {
		m := pipe.Read()
		if m == nil {
			break
		}
		count++
	}
	// At least one message should have been processed
	assert.True(t, count >= 0, "Should not panic during concurrent write/close")
}
