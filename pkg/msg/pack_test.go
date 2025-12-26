// Modified from github.com/fatedier/golib/msg/json

package msg

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sydneyowl/clh-server/msgproto"
)

var (
	TypeHandReq byte = '1'

	msgCtl = NewMsgCtl()
)

func init() {
	msgCtl.RegisterMsg(TypeHandReq, func() Message {
		return &msgproto.HandshakeRequest{}
	})
}

func TestPack(t *testing.T) {
	assert := assert.New(t)

	var msg Message
	// correct
	msg = &msgproto.HandshakeRequest{
		Os:        "Windows",
		Ver:       "0.2.1",
		AuthKey:   "d4s1c5qa1dd",
		Timestamp: 1212121212121,
		RunId:     "fsdfa",
	}
	bs, err := msgCtl.Pack(msg)
	assert.NoError(err)
	fmt.Println(len(bs))

	tp, bu, err := msgCtl.readMsg(bytes.NewReader(bs))
	assert.NoError(err)
	pack, err := msgCtl.UnPack(tp, bu)
	assert.NoError(err)
	fmt.Println(pack)
}

func TestUnPack(t *testing.T) {
	assert := assert.New(t)

	msg := &msgproto.HandshakeRequest{
		Os:        "Linux",
		Ver:       "1.0.0",
		AuthKey:   "testkey",
		Timestamp: 1234567890,
		RunId:     "testRunId",
	}

	bs, err := msgCtl.Pack(msg)
	assert.NoError(err)

	unpacked, err := msgCtl.UnPack(TypeHandReq, bs[5:]) // Skip type byte and length
	assert.NoError(err)
	unpackedMsg := unpacked.(*msgproto.HandshakeRequest)
	assert.Equal(msg.Os, unpackedMsg.Os)
	assert.Equal(msg.Ver, unpackedMsg.Ver)
	assert.Equal(msg.AuthKey, unpackedMsg.AuthKey)
	assert.Equal(msg.Timestamp, unpackedMsg.Timestamp)
	assert.Equal(msg.RunId, unpackedMsg.RunId)
}

func TestUnPackInto(t *testing.T) {
	assert := assert.New(t)

	msg := &msgproto.HandshakeRequest{
		Os:        "Linux",
		Ver:       "1.0.0",
		AuthKey:   "testkey",
		Timestamp: 1234567890,
		RunId:     "testRunId",
	}

	bs, err := msgCtl.Pack(msg)
	assert.NoError(err)

	target := &msgproto.HandshakeRequest{}
	err = msgCtl.UnPackInto(bs[5:], target) // Skip type byte and length
	assert.NoError(err)
	assert.Equal(msg.Os, target.Os)
	assert.Equal(msg.Ver, target.Ver)
	assert.Equal(msg.AuthKey, target.AuthKey)
	assert.Equal(msg.Timestamp, target.Timestamp)
	assert.Equal(msg.RunId, target.RunId)
}

func TestPackUnregisteredMessage(t *testing.T) {
	assert := assert.New(t)

	// Assuming Status is not registered
	msg := &msgproto.Status{
		DialFrequencyHz: 14074000,
		Mode:            "FT8",
	}

	_, err := msgCtl.Pack(msg)
	assert.Error(err)
	assert.Equal(ErrMsgType, err)
}

func TestUnPackInvalidType(t *testing.T) {
	assert := assert.New(t)

	buffer := []byte("invalid")
	_, err := msgCtl.UnPack('9', buffer)
	assert.Error(err)
	assert.Equal(ErrMsgType, err)
}
