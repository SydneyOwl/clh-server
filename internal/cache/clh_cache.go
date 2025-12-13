package cache

import (
	"github.com/sydneyowl/clh-server/pkg/msg"
)

//type MsgType string
//
//const (
//	TypeStatus     MsgType = "status"
//	TypeDecode     MsgType = "decode"
//	TypeWsprDecode MsgType = "wsprdecode"
//)

type CLHCache interface {
	// Add adds an message to the cache.
	Add(runID string, value msg.Message) error
	// RemoveAll deletes all elements of runID from cache
	RemoveAll(runID string) error
	// ReadAll reads all elements of runID out at once.
	ReadAll(runID string) ([]msg.Message, error)
	// ReadUntil continously read elements of runID until doneChan tells it to stop.
	ReadUntil(runID string, callback func(message msg.Message) error, doneChan <-chan struct{}) error
}
