package server

import (
	"sync"
	"time"
)

// ClientInfo holds lightweight metadata about a connected client.
type ClientInfo struct {
	RunID    string
	Type     ClientType
	Addr     string
	LastSeen time.Time
}

// ClientRegistry tracks connected clients in a thread-safe map.
type ClientRegistry struct {
	mu   sync.RWMutex
	byID map[string]ClientInfo
	// optional eviction TTL (0 = disabled)
	ttl time.Duration
}

// NewClientRegistry creates a new registry. Pass ttl>0 to enable background eviction.
func NewClientRegistry(ttl time.Duration) *ClientRegistry {
	cr := &ClientRegistry{
		byID: make(map[string]ClientInfo),
		ttl:  ttl,
	}
	if ttl > 0 {
		go cr.evictLoop()
	}
	return cr
}

// Register records or updates a client's metadata.
func (cr *ClientRegistry) Register(info ClientInfo) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	info.LastSeen = time.Now()
	cr.byID[info.RunID] = info
}

// Unregister removes a client from the registry.
func (cr *ClientRegistry) Unregister(runID string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	delete(cr.byID, runID)
}

// GetSenders returns the runIDs of clients whose type is 'sender'.
func (cr *ClientRegistry) GetSenders() []string {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	res := make([]string, 0)
	for id, info := range cr.byID {
		if info.Type == sender {
			res = append(res, id)
		}
	}
	return res
}

// Touch updates the LastSeen timestamp for a client, if present.
func (cr *ClientRegistry) Touch(runID string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if info, ok := cr.byID[runID]; ok {
		info.LastSeen = time.Now()
		cr.byID[runID] = info
	}
}

func (cr *ClientRegistry) evictLoop() {
	ticker := time.NewTicker(cr.ttl / 2)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-cr.ttl)
		cr.mu.Lock()
		for id, info := range cr.byID {
			if info.LastSeen.Before(cutoff) {
				delete(cr.byID, id)
			}
		}
		cr.mu.Unlock()
	}
}
