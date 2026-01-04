package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFull(t *testing.T) {
	// Set test values
	originalVersion := Version
	originalCommit := Commit
	defer func() {
		Version = originalVersion
		Commit = originalCommit
	}()

	Version = "1.0.0"
	Commit = "abc123"
	full := Full()
	assert.Equal(t, "1.0.0 (commit abc123)", full)
	assert.True(t, strings.Contains(full, Version))
	assert.True(t, strings.Contains(full, Commit))
}
