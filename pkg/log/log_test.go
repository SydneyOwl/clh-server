package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sydneyowl/clh-server/pkg/config"
)

func TestSetLogger(t *testing.T) {
	cfg := config.Log{
		LogToFile:        false,
		Level:            "INFO",
		LogFileDirectory: "",
	}
	err := SetLogger(cfg, false)
	assert.NoError(t, err)

	// Test invalid level
	cfg.Level = "INVALID"
	err = SetLogger(cfg, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}
