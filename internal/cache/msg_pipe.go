package cache

import (
	"sync"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

// MsgPipe is not used actually -
type MsgPipe struct {
	msgChan chan msg.Message
	closed  bool
	once    sync.Once
	mu      sync.Mutex
}

func NewMsgPipe() *MsgPipe {
	return &MsgPipe{
		msgChan: make(chan msg.Message, 20),
	}
}

func (p *MsgPipe) Close() {
	p.once.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.closed = true
		close(p.msgChan)
	})
}

func (p *MsgPipe) Read() msg.Message {
	m, ok := <-p.msgChan
	if !ok {
		return nil
	}
	return m
}

// Write is non-blocking. Drops message if buffer is full.
func (p *MsgPipe) Write(m msg.Message) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		slog.Warnf("msg pipe closed - dropping message")
		return
	}

	select {
	case p.msgChan <- m:
	default:
		slog.Warnf("msg pipe buffer full - dropping message")
	}
}
