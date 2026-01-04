package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDefaultConfig(t *testing.T) {
	cfg := GenerateDefaultConfig()
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Server)
	assert.NotNil(t, cfg.Log)
	assert.Equal(t, "0.0.0.0", cfg.Server.BindAddr)
	assert.Equal(t, 7410, cfg.Server.BindPort)
	assert.True(t, cfg.Server.Encrypt.EnableTLS)
}

func TestParseConfig(t *testing.T) {
	// Create a temporary config file
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

	cfg, err := ParseConfig(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", cfg.Server.BindAddr)
	assert.Equal(t, 8080, cfg.Server.BindPort)
	assert.Equal(t, "testkey12345", cfg.Server.Encrypt.Key)
	assert.False(t, cfg.Server.Encrypt.EnableTLS)
}

func TestConfig_Verify(t *testing.T) {
	// Valid config
	cfg := GenerateDefaultConfig()
	cfg.Server.Encrypt.Key = "validkey12345"
	warnings, errors := cfg.Verify()
	assert.Empty(t, errors)
	assert.NotEmpty(t, warnings) // TLS cert paths empty

	// Invalid IP
	cfg.Server.BindAddr = "invalid"
	warnings, errors = cfg.Verify()
	assert.Contains(t, errors, "Invalid bind address. Note that we only support IPv4 address so far!")

	// Reset
	cfg = GenerateDefaultConfig()
	cfg.Server.Encrypt.Key = "validkey12345"
	cfg.Server.BindAddr = "192.168.1.1"

	// Invalid port
	cfg.Server.BindPort = 99999
	warnings, errors = cfg.Verify()
	assert.Contains(t, errors, "Invalid bind port number.")

	// Short key
	cfg.Server.BindPort = 7410
	cfg.Server.Encrypt.Key = "short"
	warnings, errors = cfg.Verify()
	assert.Contains(t, errors, "Key must be at least 10 characters.")

	// Invalid log level
	cfg.Server.Encrypt.Key = "validkey12345"
	cfg.Log.Level = "INVALID"
	warnings, errors = cfg.Verify()
	assert.Contains(t, errors, "Provided log level is invalid. You should choose from [ PANIC,FATAL,ERROR,WARNING,NOTICE,INFO,DEBUG,TRACE ].")
}
