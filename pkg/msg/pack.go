// Modified from github.com/fatedier/golib/msg/json

package msg

import (
	"bytes"
	"encoding/binary"

	"google.golang.org/protobuf/proto"
)

func (msgCtl *MsgCtl) unpack(typeByte byte, buffer []byte, msgIn Message) (msg Message, err error) {
	if msgIn == nil {
		t, ok := msgCtl.typeMap[typeByte]
		if !ok {
			err = ErrMsgType
			return
		}

		msg = t()
	} else {
		msg = msgIn
	}

	err = proto.Unmarshal(buffer, msg)
	return
}

func (msgCtl *MsgCtl) UnPackInto(buffer []byte, msg Message) (err error) {
	_, err = msgCtl.unpack(' ', buffer, msg)
	return
}

func (msgCtl *MsgCtl) UnPack(typeByte byte, buffer []byte) (msg Message, err error) {
	return msgCtl.unpack(typeByte, buffer, nil)
}

func (msgCtl *MsgCtl) Pack(msg Message) ([]byte, error) {
	typeByte, ok := msgCtl.nameByteMap[string(proto.MessageName(msg))]
	if !ok {
		return nil, ErrMsgType
	}

	content, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	_ = buffer.WriteByte(typeByte)
	_ = binary.Write(buffer, binary.BigEndian, uint32(len(content)))
	_, _ = buffer.Write(content)
	return buffer.Bytes(), nil
}
