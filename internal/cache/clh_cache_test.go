package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

func TestMemoryCache_PublishMessage(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.RemoveCache("testRunId")

	message := &clh_proto.HandshakeRequest{
		Os:        "Linux",
		Ver:       "1.0.0",
		AuthKey:   "testkey",
		Timestamp: 1234567890,
		RunId:     "testRunId",
	}

	err := cache.PublishMessage("testRunId", message)
	assert.NoError(t, err)
}

func TestMemoryCache_SubscribeHandler(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.RemoveCache("testRunId")

	var receivedMessages []msg.Message
	var mu sync.Mutex

	handler := func(messages []msg.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, messages...)
	}

	token := cache.SubscribeHandler("testRunId", handler)
	assert.NotNil(t, token)

	message := &clh_proto.HandshakeRequest{
		Os:        "Linux",
		Ver:       "1.0.0",
		AuthKey:   "testkey",
		Timestamp: 1234567890,
		RunId:     "testRunId",
	}

	err := cache.PublishMessage("testRunId", message)
	assert.NoError(t, err)

	// Wait a bit for async processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, message, receivedMessages[0])
	mu.Unlock()

	cache.UnsubscribeHandler("testRunId", token)
}

func TestMemoryCache_BufferedMessages(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.RemoveCache("testRunId")

	message1 := &clh_proto.HandshakeRequest{
		Os:        "Linux",
		Ver:       "1.0.0",
		AuthKey:   "testkey1",
		Timestamp: 1234567890,
		RunId:     "testRunId",
	}

	message2 := &clh_proto.HandshakeRequest{
		Os:        "Windows",
		Ver:       "1.0.0",
		AuthKey:   "testkey2",
		Timestamp: 1234567891,
		RunId:     "testRunId",
	}

	// Publish before subscribing
	err := cache.PublishMessage("testRunId", message1)
	assert.NoError(t, err)
	err = cache.PublishMessage("testRunId", message2)
	assert.NoError(t, err)

	var receivedMessages []msg.Message
	var mu sync.Mutex

	handler := func(messages []msg.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, messages...)
	}

	token := cache.SubscribeHandler("testRunId", handler)
	assert.NotNil(t, token)

	// Wait for buffered messages to be sent
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Len(t, receivedMessages, 2)
	assert.Equal(t, message1, receivedMessages[0])
	assert.Equal(t, message2, receivedMessages[1])
	mu.Unlock()

	cache.UnsubscribeHandler("testRunId", token)
}

func TestMemoryCache_RemoveCache(t *testing.T) {
	cache := NewMemoryCache()

	message := &clh_proto.HandshakeRequest{
		Os:        "Linux",
		Ver:       "1.0.0",
		AuthKey:   "testkey",
		Timestamp: 1234567890,
		RunId:     "testRunId",
	}

	err := cache.PublishMessage("testRunId", message)
	assert.NoError(t, err)

	cache.RemoveCache("testRunId")

	// After removal, subscribing should not get buffered messages
	var receivedMessages []msg.Message
	var mu sync.Mutex

	handler := func(messages []msg.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, messages...)
	}

	token := cache.SubscribeHandler("testRunId", handler)
	assert.NotNil(t, token)

	// Wait
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Len(t, receivedMessages, 0)
	mu.Unlock()

	cache.UnsubscribeHandler("testRunId", token)
}

func TestMemoryCache_StatusMessage(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.RemoveCache("testRunId")

	status := &clh_proto.Status{
		DialFrequencyHz:      14074000,
		Mode:                 "FT8",
		DxCall:               "VK3",
		Report:               "-15",
		TxMode:               "FT8",
		TxEnabled:            true,
		Transmitting:         false,
		Decoding:             true,
		RxOffsetFrequencyHz:  0,
		TxOffsetFrequencyHz:  0,
		DeCall:               "VK2",
		DeGrid:               "QF22",
		DxGrid:               "QF33",
		TxWatchdog:           false,
		SubMode:              "",
		FastMode:             false,
		SpecialOperationMode: 0,
		FrequencyTolerance:   0,
		TrPeriod:             15,
		ConfigurationName:    "Default",
		TxMessage:            "",
	}

	err := cache.PublishMessage("testRunId", status)
	assert.NoError(t, err)

	var receivedMessages []msg.Message
	var mu sync.Mutex

	handler := func(messages []msg.Message) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, messages...)
	}

	token := cache.SubscribeHandler("testRunId", handler)
	assert.NotNil(t, token)

	// Wait
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, status, receivedMessages[0])
	mu.Unlock()

	cache.UnsubscribeHandler("testRunId", token)
}
