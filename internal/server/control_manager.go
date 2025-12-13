package server

import (
	"sync"
)

type ControlManager struct {
	// controls indexed by run id
	ctlsByRunID map[string]*Control

	mu sync.RWMutex
}

func NewControlManager() *ControlManager {
	return &ControlManager{
		ctlsByRunID: make(map[string]*Control),
	}
}

func (cm *ControlManager) Add(runID string, ctl *Control) (old *Control) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var ok bool
	old, ok = cm.ctlsByRunID[runID]
	if ok {
		old.Replace(ctl)
	}
	cm.ctlsByRunID[runID] = ctl
	return
}

func (cm *ControlManager) Del(runID string, ctl *Control) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if c, ok := cm.ctlsByRunID[runID]; ok && c == ctl {
		delete(cm.ctlsByRunID, runID)
	}
}

func (cm *ControlManager) GetByID(runID string) (ctl *Control, ok bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	ctl, ok = cm.ctlsByRunID[runID]
	return
}

func (cm *ControlManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, ctl := range cm.ctlsByRunID {
		_ = ctl.Close()
	}
	cm.ctlsByRunID = make(map[string]*Control)
	return nil
}
