// Modified from github.com/fatedier/golib/msg/json

package msg

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var defaultMaxMsgLength uint32 = 10240

type Message protoreflect.ProtoMessage

type MsgCtl struct {
	typeMap     map[byte]func() Message
	nameByteMap map[string]byte

	maxMsgLength uint32
}

func NewMsgCtl() *MsgCtl {
	return &MsgCtl{
		nameByteMap:  make(map[string]byte),
		typeMap:      make(map[byte]func() Message),
		maxMsgLength: defaultMaxMsgLength,
	}
}

func (msgCtl *MsgCtl) RegisterMsg(typeByte byte, createFunc func() Message) {
	msgCtl.typeMap[typeByte] = createFunc
	msgCtl.nameByteMap[string(proto.MessageName(createFunc()))] = typeByte
}

func (msgCtl *MsgCtl) SetMaxMsgLength(length uint32) {
	msgCtl.maxMsgLength = length
}
