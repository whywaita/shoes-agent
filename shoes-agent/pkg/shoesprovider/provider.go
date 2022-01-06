package shoesprovider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-plugin"
	myshoespb "github.com/whywaita/myshoes/api/proto"
	agentpb "github.com/whywaita/shoes-agent/proto.go"
	"github.com/whywaita/shoes-agent/shoes-agent/pkg/agent"
	"github.com/whywaita/shoes-agent/shoes-agent/pkg/backend"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AgentPlugin is plugin for shoes-agent
type AgentPlugin struct {
	plugin.Plugin

	Backend   backend.Backend
	ShoesType string
}

// GRPCServer is implement gRPC Server.
func (a *AgentPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	client := NewAgentClient(a.Backend, a.ShoesType)
	myshoespb.RegisterShoesServer(s, client)
	return nil
}

// GRPCClient is implement gRPC client.
// This function is not have client, so return nil
func (a *AgentPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, nil
}

// AgentClient is a client of agent
type AgentClient struct {
	Backend   backend.Backend
	ShoesType string

	myshoespb.UnimplementedShoesServer
}

func NewAgentClient(backend backend.Backend, shoesType string) *AgentClient {
	return &AgentClient{
		Backend:   backend,
		ShoesType: shoesType,
	}
}

var (
	// ErrAcceptableAgentNotFound is error message of "acceptable agent is not found"
	ErrAcceptableAgentNotFound = fmt.Errorf("acceptable agent is not found")
	// ErrAlreadyCreated is error message for already created
	ErrAlreadyCreated = fmt.Errorf("already created")
)

func (a *AgentClient) schedule(ctx context.Context) (*backend.Agent, error) {
	retryCount := 60

	var target *backend.Agent

	for i := 0; i <= retryCount; i++ {
		time.Sleep(1 * time.Second)
		log.Printf("waiting start shoes-agent... [%d/%d]\n", i, retryCount)

		agents, err := a.Backend.ListAgent(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve list of agent: %w", err)
		}
		log.Printf("found agents: %+v\n", agents)

		t, err := scheduleAgent(agents)
		if err != nil {
			if isRetryableError(err) {
				log.Printf("found retryable error, will retry: %+v\n", err)
				continue
			}
			return nil, fmt.Errorf("failed to schedule agent: %w", err)
		}

		target = t
		break
	}
	if target == nil {
		return nil, ErrAcceptableAgentNotFound
	}

	return target, nil
}

func scheduleAgent(agents []backend.Agent) (*backend.Agent, error) {
	var acceptable []backend.Agent

	for _, a := range agents {
		if a.Status == agentpb.Status_Booting {
			acceptable = append(acceptable, a)
		}
	}

	if len(acceptable) == 0 {
		return nil, ErrAcceptableAgentNotFound
	}

	// scheduling algorithm
	target := acceptable[0]

	return &target, nil
}

func isRetryableError(err error) bool {
	switch {
	case errors.Is(err, ErrAcceptableAgentNotFound):
		return true
	}
	return false
}

func (a *AgentClient) AddInstance(ctx context.Context, req *myshoespb.AddInstanceRequest) (*myshoespb.AddInstanceResponse, error) {
	err := a.Backend.CreateInstance(ctx, req.RunnerName)
	if err != nil && !errors.Is(err, ErrAlreadyCreated) {
		return nil, status.Errorf(codes.Internal, "failed to create an instance: %+v", err)
	}

	target, err := a.schedule(ctx)
	if err != nil {
		switch {
		case errors.Is(err, ErrAcceptableAgentNotFound):
			return nil, status.Errorf(codes.NotFound, "failed to schedule agent: %+v", err)
		default:
			return nil, status.Errorf(codes.Internal, "failed to schedule agent: %+v", err)
		}
	}

	log.Printf("target host is found! (host: %+v)\n", target)

	c, err := agent.NewClient(target.GRPCHost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to agent.NewClient: %+v", err)
	}
	client := agentpb.NewShoesAgentClient(c.Conn)

	if _, err := client.StartRunner(ctx, &agentpb.StartRunnerRequest{
		SetupScript: req.SetupScript,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start runner: %+v", err)
	}

	return &myshoespb.AddInstanceResponse{
		CloudId:   target.CloudID,
		ShoesType: a.ShoesType,
		IpAddress: target.GRPCHost,
	}, nil
}

func (a *AgentClient) DeleteInstance(ctx context.Context, req *myshoespb.DeleteInstanceRequest) (*myshoespb.DeleteInstanceResponse, error) {
	if _, err := a.Backend.GetAgent(ctx, req.CloudId); err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get agent: %+v", err)
	}

	if err := a.Backend.DeleteInstance(ctx, req.CloudId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete instance from backend: %+v", err)
	}

	return &myshoespb.DeleteInstanceResponse{}, nil
}
