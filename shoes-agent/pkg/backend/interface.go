package backend

import (
	"context"

	agentpb "github.com/whywaita/shoes-agent/proto.go"
)

// Backend is backend infrastructure
type Backend interface {
	// GetAgent retrieve an agent
	GetAgent(ctx context.Context, cloudID string) (*Agent, error)
	// ListAgent retrieve a list of all agent
	ListAgent(ctx context.Context) ([]Agent, error)
	// CreateInstance create an instance that running agent in backend
	CreateInstance(ctx context.Context, runnerName string) error
	// DeleteInstance delete an instance
	DeleteInstance(ctx context.Context, cloudID string) error
}

// Agent is instance of shoes-agent
type Agent struct {
	CloudID  string
	GRPCHost string
	Status   agentpb.Status
}
