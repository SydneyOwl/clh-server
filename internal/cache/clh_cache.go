package cache

import (
	"context"
	"github.com/sydneyowl/clh-server/pkg/msg"
)

type CLHCache interface {
	Add(runID string, value msg.Message, ctx context.Context) error
	RemoveAll(runID string, ctx context.Context) error
	ReadAll(runID string, ctx context.Context) ([]msg.Message, error)
	ReadAllType(runID string, msgType string, ctx context.Context) ([]msg.Message, error)
	StartRead(runID string, callback func(message msg.Message), ctx context.Context) error
}
