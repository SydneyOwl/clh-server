package msg

import (
	"github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

func buildMsgMap() map[byte]func() msg.Message {
	var msgMap = make(map[byte]func() msg.Message)
	msgMap['1'] = func() msg.Message { return &clh_proto.Ping{} }
	msgMap['2'] = func() msg.Message { return &clh_proto.Pong{} }
	msgMap['3'] = func() msg.Message { return &clh_proto.Command{} }
	msgMap['4'] = func() msg.Message { return &clh_proto.CommonResponse{} }
	msgMap['5'] = func() msg.Message { return &clh_proto.CommandResponse{} }
	msgMap['a'] = func() msg.Message { return &clh_proto.HandshakeRequest{} }
	msgMap['b'] = func() msg.Message { return &clh_proto.HandshakeResponse{} }
	msgMap['c'] = func() msg.Message { return &clh_proto.WsjtxMessage{} }
	msgMap['d'] = func() msg.Message { return &clh_proto.PackedMessage{} }
	msgMap['e'] = func() msg.Message { return &clh_proto.RigData{} }
	return msgMap
}
