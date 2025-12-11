package verifier

import "github.com/sydneyowl/clh-server/msgproto"

type Verifier interface {
	VerifyLogin(*msgproto.HandshakeRequest) error
}
