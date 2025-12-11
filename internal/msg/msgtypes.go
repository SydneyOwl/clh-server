package msg

import (
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

func buildMsgMap() map[byte]func() msg.Message {
	var msgMap = make(map[byte]func() msg.Message)
	msgMap['a'] = func() msg.Message { return &msgproto.HandshakeRequest{} }
	msgMap['b'] = func() msg.Message { return &msgproto.HandshakeResponse{} }
	msgMap['c'] = func() msg.Message { return &msgproto.DigModeUpdate{} }
	return msgMap
}
