package test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	cached1 "github.com/sydneyowl/clh-server/internal/cache"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

// mock message
var testMsg1 = &msgproto.WsjtxMessage{}
var testMsg2 = &msgproto.WsjtxMessage{}
var testMsg3 = &msgproto.WsjtxMessage{}
var testMsg4 = &msgproto.WsjtxMessage{}

func TestMemoryCache_PublishAndSubscribe(t *testing.T) {
	cache := cached1.NewMemoryCache()
	runId := "test-run-1"

	err := cache.PublishMessage(runId, testMsg1)
	assert.NoError(t, err)
	err = cache.PublishMessage(runId, testMsg2)
	assert.NoError(t, err)

	// 3. 注册 subscriber
	var received []msg.Message
	var mu sync.Mutex
	handler := func(m []msg.Message) {
		mu.Lock()
		received = append(received, m...)
		mu.Unlock()
	}

	ch := cache.SubscribeHandler(runId, handler)

	// 4. 等待 emitter 处理（因为是同步调用，无需 wait）
	time.Sleep(10 * time.Millisecond)

	// 5. 验证历史消息已回放
	mu.Lock()
	assert.Equal(t, []msg.Message{testMsg1, testMsg2}, received)
	mu.Unlock()

	// 7. 发新消息，应直接投递
	err = cache.PublishMessage(runId, testMsg3)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	assert.Equal(t, []msg.Message{testMsg1, testMsg2, testMsg3}, received)
	mu.Unlock()

	// 8. 取消订阅
	cache.UnsubscribeHandler(runId, ch)

	// 9. 再发消息 → 应重新缓存（因为无 listener）
	err = cache.PublishMessage(runId, testMsg4)
	assert.NoError(t, err)
}
