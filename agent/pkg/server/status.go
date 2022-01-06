package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/whywaita/shoes-agent/proto.go"
)

const (
	FileStatus = "shoes-agent.status"
	FileScript = "shoes-agent.sh"
)

func marshalStatus(agentStatus pb.Status) string {
	return agentStatus.String()
}

func unmarshalStatus(in string) pb.Status {
	switch in {
	case "Unknown":
		return pb.Status_Unknown
	case "Booting":
		return pb.Status_Booting
	case "Idle":
		return pb.Status_Idle
	case "Active":
		return pb.Status_Active
	case "Offline":
		return pb.Status_Offline
	default:
		return pb.Status_Unknown
	}
}

// GetAgentStatus response a status of agent
func (s *AgentServer) GetAgentStatus(ctx context.Context, req *pb.GetAgentStatusRequest) (*pb.GetAgentStatusResponse, error) {
	agentStatus, err := s.getAgentStatus()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve status of agent from file: %+v", err)
	}

	return &pb.GetAgentStatusResponse{
		Status: agentStatus,
	}, nil
}

func getStatusFilePath() string {
	return filepath.Join(PathStatus, FileStatus)
}

func getScriptFilePath() string {
	return filepath.Join(PathStatus, FileScript)
}

// getAgentStatus get status of agent from status file
func (s *AgentServer) getAgentStatus() (pb.Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out, err := os.ReadFile(getStatusFilePath())
	if err != nil {
		return pb.Status_Unknown, fmt.Errorf("failed to read status file: %w", err)
	}

	agentStatus := unmarshalStatus(string(out))
	return agentStatus, nil
}

// setAgentStatus set status to status file
func (s *AgentServer) setAgentStatus(status pb.Status) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Create(getStatusFilePath())
	if err != nil {
		return fmt.Errorf("failed to open status file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(marshalStatus(status)); err != nil {
		return fmt.Errorf("failed to write status file: %w", err)
	}
	return nil
}
