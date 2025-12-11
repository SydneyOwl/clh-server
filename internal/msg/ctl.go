// Modified version of github.com/fatedier/frp/pkg/msg

package msg

import (
	"io"

	"github.com/gookit/slog"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

var msgCtl *msg.MsgCtl

func init() {
	msgCtl = msg.NewMsgCtl()

	for typeByte, ctor := range buildMsgMap() {
		slog.Tracef("Registering msgtypes %s", typeByte)
		msgCtl.RegisterMsg(typeByte, ctor)
	}
}

func ReadMsg(c io.Reader) (msg msg.Message, err error) {
	return msgCtl.ReadMsg(c)
}

func ReadMsgInto(c io.Reader, msg msg.Message) (err error) {
	return msgCtl.ReadMsgInto(c, msg)
}

func WriteMsg(c io.Writer, msg msg.Message) (err error) {
	return msgCtl.WriteMsg(c, msg)
}
