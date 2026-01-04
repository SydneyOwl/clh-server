package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateClientType(t *testing.T) {
	ct, err := ValidateClientType("receiver")
	assert.NoError(t, err)
	assert.Equal(t, ClientType("receiver"), ct)

	ct, err = ValidateClientType("sender")
	assert.NoError(t, err)
	assert.Equal(t, ClientType("sender"), ct)

	ct, err = ValidateClientType("connTest")
	assert.NoError(t, err)
	assert.Equal(t, ClientType("connTest"), ct)

	_, err = ValidateClientType("invalid")
	assert.Error(t, err)
}
