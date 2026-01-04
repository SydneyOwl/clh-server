package verifier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clh_proto "github.com/sydneyowl/clh-server/clh-proto"
	"github.com/sydneyowl/clh-server/pkg/crypto"
)

func TestAuthKeyVerifier_VerifyLogin(t *testing.T) {
	key := "testkey"
	verifier := NewAuthKeyVerifier(key)

	tm := time.Now().Unix()
	authKey := crypto.CalcAuthKey(key, tm)

	req := &clh_proto.HandshakeRequest{
		RunId:     "test",
		Timestamp: tm,
		AuthKey:   authKey,
	}

	err := verifier.VerifyLogin(req)
	assert.NoError(t, err)

	// Invalid run id
	req.RunId = ""
	err = verifier.VerifyLogin(req)
	assert.Error(t, err)

	// Invalid timestamp
	req.RunId = "test"
	req.Timestamp = time.Now().Unix() - 20
	err = verifier.VerifyLogin(req)
	assert.Error(t, err)

	// Invalid key
	req.Timestamp = tm
	req.AuthKey = "wrong"
	err = verifier.VerifyLogin(req)
	assert.Error(t, err)
}
