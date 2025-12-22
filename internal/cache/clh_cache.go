package cache

import (
	"github.com/sydneyowl/clh-server/pkg/msg"
)

// CLHCache defines a message caching and dispatching interface with buffered replay capability.
type CLHCache interface {
	// PublishMessage publishes a message to the topic associated with the given runId
	PublishMessage(runId string, message msg.Message) error

	// SubscribeHandler registers a message handler for the given runId.
	// The returned token must be used to unsubscribe via UnsubscribeHandler.
	SubscribeHandler(runId string, handler func(message []msg.Message)) (token any)

	// UnsubscribeHandler removes the subscription for the given runId.
	// The token must be the value previously returned by SubscribeHandler.
	UnsubscribeHandler(runId string, token any)

	// RemoveCache removes all related cache of runId.
	RemoveCache(runId string)

	// GetSenderList gets available sender names.
	//GetSenderList() []string
}
