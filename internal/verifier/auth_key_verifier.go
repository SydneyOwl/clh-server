package verifier

import (
	"crypto/subtle"
	"errors"
	"math"
	"time"

	"github.com/sydneyowl/clh-server/msgproto"
	"github.com/sydneyowl/clh-server/pkg/crypto"
)

type AuthKeyVerifier struct {
	key string
}

func (akv *AuthKeyVerifier) VerifyLogin(m *msgproto.HandshakeRequest) error {
	if m.RunId == "" {
		return errors.New("no run id provided")
	}
	// verify if key matches using our private key
	if math.Abs(float64(m.Timestamp-time.Now().Unix())) > 5 {
		return errors.New("invalid timestamp")
	}

	calKey := crypto.CalcAuthKey(akv.key, m.Timestamp)
	if subtle.ConstantTimeCompare([]byte(calKey), []byte(m.AuthKey)) != 1 {
		return errors.New("invalid auth key")
	}
	return nil
}

func NewAuthKeyVerifier(key string) *AuthKeyVerifier {
	return &AuthKeyVerifier{
		key: key,
	}
}
