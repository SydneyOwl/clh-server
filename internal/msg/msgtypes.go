package msg

import (
	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

func buildMsgMap() map[byte]func() msg.Message {
	var msgMap = make(map[byte]func() msg.Message)
	msgMap['1'] = func() msg.Message { return &msgproto.Ping{} }
	msgMap['2'] = func() msg.Message { return &msgproto.Pong{} }
	msgMap['3'] = func() msg.Message { return &msgproto.Command{} }
	msgMap['4'] = func() msg.Message { return &msgproto.CommonResponse{} }
	msgMap['5'] = func() msg.Message { return &msgproto.CommandResponse{} }
	msgMap['a'] = func() msg.Message { return &msgproto.HandshakeRequest{} }
	msgMap['b'] = func() msg.Message { return &msgproto.HandshakeResponse{} }
	msgMap['c'] = func() msg.Message { return &msgproto.WsjtxMessage{} }
	msgMap['d'] = func() msg.Message { return &msgproto.WsjtxMessagePacked{} }
	return msgMap
}
