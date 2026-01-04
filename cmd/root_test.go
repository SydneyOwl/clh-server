package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCfg(t *testing.T) {
	// Create a temp config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
server:
  bind_addr: "127.0.0.1"
  bind_port: 8080
  encrypt:
    key: "testkey12345"
    encrypt: false
log:
  level: "INFO"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	assert.NoError(t, err)

	cfg, err := ParseCfg(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", cfg.Server.BindAddr)
	assert.Equal(t, 8080, cfg.Server.BindPort)
}
