package msg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	clh_proto "github.com/sydneyowl/clh-server/clh-proto"
)

func TestBuildMsgMap(t *testing.T) {
	msgMap := buildMsgMap()
	assert.NotNil(t, msgMap)

	// Test some entries
	ping := msgMap['1']()
	assert.IsType(t, &clh_proto.Ping{}, ping)

	pong := msgMap['2']()
	assert.IsType(t, &clh_proto.Pong{}, pong)

	handshakeReq := msgMap['a']()
	assert.IsType(t, &clh_proto.HandshakeRequest{}, handshakeReq)
}
