package trans

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomPrivateKey(t *testing.T) {
	keyPEM, err := NewRandomPrivateKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, keyPEM)

	// Decode and check it's a valid RSA private key
	block, _ := pem.Decode(keyPEM)
	assert.NotNil(t, block)
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)

	_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	assert.NoError(t, err)
}

func TestNewServerTLSConfig(t *testing.T) {
	// Test with random cert
	config, err := NewServerTLSConfig("", "", "")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Certificates, 1)
}

func TestNewClientTLSConfig(t *testing.T) {
	config, err := NewClientTLSConfig("", "", "", "example.com")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "example.com", config.ServerName)
	assert.True(t, config.InsecureSkipVerify)
}

func TestNewCertPool(t *testing.T) {
	// Create a temp CA file
	tempDir := t.TempDir()
	caFile := filepath.Join(tempDir, "ca.pem")

	// Generate a random cert as CA
	cert, err := newRandomTLSKeyPair()
	assert.NoError(t, err)

	// Write the cert PEM to file
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]})
	err = os.WriteFile(caFile, certPEM, 0644)
	assert.NoError(t, err)

	pool, err := newCertPool(caFile)
	assert.NoError(t, err)
	assert.NotNil(t, pool)
}

func TestNewCustomTLSKeyPair(t *testing.T) {
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")

	// Generate a random key pair
	cert, err := newRandomTLSKeyPair()
	assert.NoError(t, err)

	// Get the cert PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]})

	// Get the key PEM from the cert's private key
	key, ok := cert.PrivateKey.(*rsa.PrivateKey)
	assert.True(t, ok)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	err = os.WriteFile(certFile, certPEM, 0644)
	assert.NoError(t, err)
	err = os.WriteFile(keyFile, keyPEM, 0644)
	assert.NoError(t, err)

	loadedCert, err := newCustomTLSKeyPair(certFile, keyFile)
	assert.NoError(t, err)
	assert.NotNil(t, loadedCert)
}
