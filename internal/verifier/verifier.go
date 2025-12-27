package verifier

import "github.com/sydneyowl/clh-server/clh-proto"

type Verifier interface {
	VerifyLogin(*clh_proto.HandshakeRequest) error
}
