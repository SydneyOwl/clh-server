package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/olebedev/emitter"
	"github.com/patrickmn/go-cache"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

const (
	ChannelBufferSize    = 100
	BufferSliceCap       = 50
	UnsendExpireDuration = time.Minute * 30
	UnsendCleanDuration  = time.Minute * 35
	UnsendSizeLimit      = 200
	EmitTopicPrefix      = "clhmsg"
)

type TopicBuffer struct {
	mu   sync.Mutex
	msgs []msg.Message
}

type MemoryCache struct {
	unsendCache *cache.Cache
	msgEmitter  *emitter.Emitter

	// mu locks unsendCache
	mu sync.Mutex
}

func (c *MemoryCache) RemoveCache(runId string) {
	topic := fmt.Sprintf("%s:%s", EmitTopicPrefix, runId)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgEmitter.Off(topic)
	c.unsendCache.Delete(topic)
}

func (c *MemoryCache) PublishMessage(runId string, message msg.Message) error {
	return c.publishMessage(runId, message)
}

func (c *MemoryCache) SubscribeHandler(runId string, handler func(message msg.Message)) (token any) {
	wrapIt := c.eventHandlerWrapper(handler)
	return c.subscribeHandler(runId, wrapIt)
}

func (c *MemoryCache) UnsubscribeHandler(runId string, token any) {
	c.unsubscribeHandler(runId, token.(<-chan emitter.Event))
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		unsendCache: cache.New(UnsendExpireDuration, UnsendCleanDuration),
		msgEmitter:  emitter.New(ChannelBufferSize),
	}
}

func (c *MemoryCache) publishMessage(runId string, message msg.Message) error {
	topic := fmt.Sprintf("%s:%s", EmitTopicPrefix, runId)

	if len(c.msgEmitter.Listeners(topic)) != 0 {
		_ = c.msgEmitter.Emit(topic, message)
		return nil
	}

	// there's no handler that subscribes this topic - we just cache it for future usage.
	cc, found := c.unsendCache.Get(topic)
	if !found {
		t := &TopicBuffer{
			mu:   sync.Mutex{},
			msgs: make([]msg.Message, 0, BufferSliceCap),
		}
		t.mu.Lock()
		t.msgs = append(t.msgs, message)
		t.mu.Unlock()
		c.unsendCache.Set(topic, t, cache.DefaultExpiration)
	} else {
		mm, ok := cc.(*TopicBuffer)
		if !ok {
			return fmt.Errorf("unexpected type %T", cc)
		}
		// todo this is not elegant ..
		mm.mu.Lock()
		if len(mm.msgs) >= UnsendSizeLimit {
			mm.msgs = mm.msgs[1:]
		}
		mm.msgs = append(mm.msgs, message)
		mm.mu.Unlock()
		c.unsendCache.Set(topic, mm, cache.DefaultExpiration)
	}
	return nil
}

func (c *MemoryCache) subscribeHandler(runId string, handler func(event *emitter.Event)) <-chan emitter.Event {
	topic := fmt.Sprintf("%s:%s", EmitTopicPrefix, runId)
	ch := c.msgEmitter.On(topic, handler)
	c.mu.Lock()
	defer c.mu.Unlock()
	// if we're the first subscriber... send all buffered messages!
	if cc, found := c.unsendCache.Get(topic); found {
		buf := cc.(*TopicBuffer)
		buf.mu.Lock()
		history := make([]any, len(buf.msgs), BufferSliceCap)
		for i := range buf.msgs {
			history[i] = buf.msgs[i]
		}
		buf.msgs = buf.msgs[:0]
		buf.mu.Unlock()
		// delete cache...
		c.unsendCache.Delete(topic)
		if len(history) > 0 {
			c.msgEmitter.Emit(topic, history...)
		}
	}
	return ch
}

func (c *MemoryCache) unsubscribeHandler(runId string, handler <-chan emitter.Event) {
	topic := fmt.Sprintf("%s:%s", EmitTopicPrefix, runId)
	c.msgEmitter.Off(topic, handler)
}

func (c *MemoryCache) eventHandlerWrapper(msgFunc func(message msg.Message)) func(event *emitter.Event) {
	return func(event *emitter.Event) {
		for _, v := range event.Args {
			message, ok := v.(msg.Message)
			if !ok {
				slog.Debugf("Seems like emitted arg (%s) is not a protomsg....", v)
				continue
				//return
			}
			msgFunc(message)
		}
	}
}
