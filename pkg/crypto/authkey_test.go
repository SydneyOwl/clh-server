package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcAuthKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		timestamp int64
		expected  string
	}{
		{
			name:      "basic test",
			key:       "testkey",
			timestamp: 1234567890,
			expected:  "dc93e9df2e2de4ff74c76528ca9f85be4d9b80ea15492ff5addd0ee359847c29",
		},
		{
			name:      "empty key",
			key:       "",
			timestamp: 0,
			expected:  "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9",
		},
		{
			name:      "different timestamp",
			key:       "key",
			timestamp: 987654321,
			expected:  "a2acf335165a6cb22db7ecad2b56d6c7b229c196c4e13b67535dcd03571df601",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcAuthKey(tt.key, tt.timestamp)
			assert.Equal(t, tt.expected, result)
		})
	}
}
