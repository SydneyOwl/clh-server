package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

const (
	StorageTimeoutSec   = 86400
	CheckPollTimeoutSec = 60
	ChannelBufferSize   = 100
)

type MemoryCache struct {
	persistent map[string]chan msg.Message
	updateAt   map[string]time.Time
	mu         sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		persistent: make(map[string]chan msg.Message),
		updateAt:   make(map[string]time.Time),
	}
	go mc.timeoutScanner()
	return mc
}

func (m *MemoryCache) timeoutScanner() {
	ticker := time.NewTicker(CheckPollTimeoutSec * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		m.cleanupExpired()
	}
}

func (m *MemoryCache) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expiredRuns := make([]string, 0)

	for runID := range m.persistent {
		lastUpdate, exists := m.updateAt[runID]
		if !exists {
			continue
		}

		if now.Sub(lastUpdate) > StorageTimeoutSec*time.Second {
			expiredRuns = append(expiredRuns, runID)
		}
	}

	for _, runID := range expiredRuns {
		delete(m.persistent, runID)
		delete(m.updateAt, runID)
		slog.Debugf("Cleaned up expired cache for run: %s", runID)
	}
}

func (m *MemoryCache) ensureAdd(runID string, message msg.Message) error {
	m.mu.Lock()

	ch, exists := m.persistent[runID]
	if !exists {
		ch = make(chan msg.Message, ChannelBufferSize)
		m.persistent[runID] = ch
	}

	m.updateAt[runID] = time.Now()
	m.mu.Unlock()

	// 尝试发送消息
	select {
	case ch <- message:
		return nil
	default:
		select {
		case <-ch:
			slog.Tracef("Channel full for %s, dropping oldest message", runID)
		default:
		}

		select {
		case ch <- message:
			return nil
		default:
			slog.Warnf("Channel for run %s is congested, message dropped", runID)
			return fmt.Errorf("channel for run %s is congested", runID)
		}
	}
}

func (m *MemoryCache) getChannel(runID string) (chan msg.Message, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ch, exists := m.persistent[runID]
	return ch, exists
}

func (m *MemoryCache) Add(runID string, value msg.Message) error {
	switch x := value.(type) {
	case *msgproto.WsjtxMessage:
		switch v := x.Payload.(type) {
		case *msgproto.WsjtxMessage_Status:
			return m.ensureAdd(runID, v.Status)
		case *msgproto.WsjtxMessage_Decode:
			return m.ensureAdd(runID, v.Decode)
		case *msgproto.WsjtxMessage_WsprDecode:
			return m.ensureAdd(runID, v.WsprDecode)
		default:
			return fmt.Errorf("unsupported message type: %T", v)
		}
	default:
		return fmt.Errorf("unsupported message type: %T", value)
	}
}

func (m *MemoryCache) RemoveAll(runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ch, exists := m.persistent[runID]; exists {
		close(ch)
	}
	delete(m.persistent, runID)
	delete(m.updateAt, runID)

	return nil
}

func (m *MemoryCache) ReadAll(runID string) ([]msg.Message, error) {
	ch, exists := m.getChannel(runID)
	if !exists {
		return nil, fmt.Errorf("%s is not cached", runID)
	}

	msgs := make([]msg.Message, 0, 30)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return msgs, nil
			}
			msgs = append(msgs, msg)
		default:
			return msgs, nil
		}
	}
}

func (m *MemoryCache) ReadUntil(runID string, callback func(message msg.Message) error, doneChan <-chan struct{}) error {
	ch, exists := m.getChannel(runID)
	if !exists {
		return fmt.Errorf("cache for run %s not found", runID)
	}

	go func() {
		for {
			select {
			case <-doneChan:
				slog.Debugf("Stop reading from cache of %s", runID)
				return

			case msg, ok := <-ch:
				if !ok {
					slog.Debugf("channel of %s closed", runID)
					return
				}

				if err := callback(msg); err != nil {
					slog.Errorf("Callback error for %s: %v", runID, err)
				}
			}
		}
	}()

	return nil
}
