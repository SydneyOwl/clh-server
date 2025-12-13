package cache

import (
	"fmt"
	"sync"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

type MemoryCache struct {
	// runid - msgtype - message
	storage map[string]chan msg.Message
	mu      sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		storage: make(map[string]chan msg.Message),
	}
}

func (m *MemoryCache) ensureLaneExist(runID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.storage[runID]
	if !ok {
		m.storage[runID] = make(chan msg.Message, 100)
	}
}

func (m *MemoryCache) Add(runID string, value msg.Message) error {
	switch x := value.(type) {
	case *msgproto.WsjtxMessage:
		switch v := x.Payload.(type) {
		case *msgproto.WsjtxMessage_Status:
			m.ensureLaneExist(runID)
			m.storage[runID] <- v.Status
		case *msgproto.WsjtxMessage_Decode:
			m.ensureLaneExist(runID)
			m.storage[runID] <- v.Decode
		case *msgproto.WsjtxMessage_WsprDecode:
			m.ensureLaneExist(runID)
			m.storage[runID] <- v.WsprDecode
		default:
			return fmt.Errorf("Received unexpected wsjtx message type: %T\n", v)
		}
	default:
		return fmt.Errorf("received unexpected message type: %T", value)
	}
	return nil
}

func (m *MemoryCache) RemoveAll(runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.storage, runID)
	return nil
}

func (m *MemoryCache) ReadAll(runID string) ([]msg.Message, error) {
	msgs := make([]msg.Message, 0, 30)
	rc, exists := m.storage[runID]
	if !exists {
		return msgs, nil
	}
	for {
		select {
		case rr, ok := <-rc:
			if !ok {
				return msgs, nil
			}
			msgs = append(msgs, rr)
		default:
			return msgs, nil
		}
	}
}

func (m *MemoryCache) ReadUntil(runID string, callback func(message msg.Message), doneChan <-chan struct{}) error {
	rc, exists := m.storage[runID]
	if !exists {
		return fmt.Errorf("cache for run %s not found", runID)
	}

	go func(callback func(message msg.Message), doneChan <-chan struct{}) {
		for {
			select {
			case <-doneChan:
				slog.Debugf("Stop reading from cache of %s", runID)
				return
			case q, ok := <-rc:
				if !ok {
					slog.Debugf("channel of %s closed. stop reading.", runID)
					return
				}
				go callback(q)
			}
		}
	}(callback, doneChan)
	return nil
}
