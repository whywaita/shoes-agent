package mock

import (
	"context"

	agentpb "github.com/whywaita/shoes-agent/proto.go"
	"github.com/whywaita/shoes-agent/shoes-agent/pkg/backend"
)

// Mock is mock of backend infrastructure
type Mock struct{}

// New create a mock
func New() *Mock {
	return &Mock{}
}

func (m *Mock) GetAgent(ctx context.Context, cloudID string) (*backend.Agent, error) {
	return &backend.Agent{
		CloudID:  cloudID,
		GRPCHost: "192.0.2.1",
		Status:   agentpb.Status_Idle,
	}, nil
}

func (m *Mock) ListAgent(ctx context.Context) ([]backend.Agent, error) {
	return []backend.Agent{
		{
			CloudID:  "cloud-id-192-0-2-1",
			GRPCHost: "192.0.2.1",
			Status:   agentpb.Status_Idle,
		},
	}, nil
}

func (m Mock) CreateInstance(ctx context.Context) error {
	return nil
}

func (m *Mock) DeleteInstance(ctx context.Context, cloudID string) error {
	return nil
}
