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
