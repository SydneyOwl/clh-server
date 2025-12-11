// Modified version of github.com/fatedier/frp/pkg/msg

package msg

import (
	"io"

	"github.com/sydneyowl/clh-server/pkg/msg"
	"google.golang.org/protobuf/proto"
)

func AsyncHandler(f func(msg.Message)) func(msg.Message) {
	return func(m msg.Message) {
		go f(m)
	}
}

// Dispatcher is used to send messages to net.Conn or register handlers for messages read from net.Conn.
type Dispatcher struct {
	rw io.ReadWriter

	sendCh         chan msg.Message
	doneCh         chan struct{}
	msgHandlers    map[string]func(msg.Message)
	defaultHandler func(msg.Message)
}

func NewDispatcher(rw io.ReadWriter) *Dispatcher {
	return &Dispatcher{
		rw:          rw,
		sendCh:      make(chan msg.Message, 100),
		doneCh:      make(chan struct{}),
		msgHandlers: make(map[string]func(msg.Message)),
	}
}

// Run will block until io.EOF or some error occurs.
func (d *Dispatcher) Run() {
	go d.sendLoop()
	go d.readLoop()
}

func (d *Dispatcher) sendLoop() {
	for {
		select {
		case <-d.doneCh:
			return
		case m := <-d.sendCh:
			_ = WriteMsg(d.rw, m)
		}
	}
}

func (d *Dispatcher) readLoop() {
	for {
		m, err := ReadMsg(d.rw)
		if err != nil {
			close(d.doneCh)
			return
		}

		if handler, ok := d.msgHandlers[string(proto.MessageName(m))]; ok {
			handler(m)
		} else if d.defaultHandler != nil {
			d.defaultHandler(m)
		}
	}
}

func (d *Dispatcher) Send(m msg.Message) error {
	select {
	case <-d.doneCh:
		return io.EOF
	case d.sendCh <- m:
		return nil
	}
}

func (d *Dispatcher) RegisterHandler(msg msg.Message, handler func(msg.Message)) {
	d.msgHandlers[string(proto.MessageName(msg))] = handler
}

func (d *Dispatcher) RegisterDefaultHandler(handler func(msg.Message)) {
	d.defaultHandler = handler
}

func (d *Dispatcher) Done() chan struct{} {
	return d.doneCh
}
