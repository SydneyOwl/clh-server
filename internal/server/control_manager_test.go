package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControlManager_Add(t *testing.T) {
	cm := NewControlManager()
	ctl1 := &Control{}
	ctl2 := &Control{}

	old := cm.Add("run1", ctl1)
	assert.Nil(t, old)

	// Don't add same runID to avoid Replace
	old = cm.Add("run2", ctl2)
	assert.Nil(t, old)

	ctl, ok := cm.GetByID("run1")
	assert.True(t, ok)
	assert.Equal(t, ctl1, ctl)
}

func TestControlManager_Del(t *testing.T) {
	cm := NewControlManager()
	ctl := &Control{}
	cm.Add("run1", ctl)

	ctl2 := &Control{}
	cm.Del("run1", ctl2) // wrong ctl, should not delete

	_, ok := cm.GetByID("run1")
	assert.True(t, ok)

	cm.Del("run1", ctl)
	_, ok = cm.GetByID("run1")
	assert.False(t, ok)
}

func TestControlManager_GetByID(t *testing.T) {
	cm := NewControlManager()
	ctl := &Control{}
	cm.Add("run1", ctl)

	retrieved, ok := cm.GetByID("run1")
	assert.True(t, ok)
	assert.Equal(t, ctl, retrieved)

	_, ok = cm.GetByID("nonexistent")
	assert.False(t, ok)
}
