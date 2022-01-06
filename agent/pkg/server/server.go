package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	pb "github.com/whywaita/shoes-agent/proto.go"
	"google.golang.org/grpc"
)

const (
	// AgentListenPort is listen port of agent gRPC server
	AgentListenPort = 5676
)

// AgentServer implement gRPC server
type AgentServer struct {
	pb.UnimplementedShoesAgentServer

	mu sync.Mutex
}

// New create gRPC server
func New() *AgentServer {
	as := &AgentServer{
		mu: sync.Mutex{},
	}

	go func() {
		if err := as.pollingRunnerProcess(); err != nil {
			log.Printf("failed to polling runner process: %+v\n", err)
			return
		}
	}()

	return as
}

// Run start gRPC server
func (s *AgentServer) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", AgentListenPort))
	if err != nil {
		return fmt.Errorf("failed to listen port (port: %d): %w", AgentListenPort, err)
	}
	log.Printf("listen port: %d\n", AgentListenPort)

	grpcServer := grpc.NewServer()
	pb.RegisterShoesAgentServer(grpcServer, s)

	if err := s.setAgentStatus(pb.Status_Booting); err != nil {
		return fmt.Errorf("failed to set agent status (status: booting): %w", err)
	}
	defer func() {
		if err := os.Remove(getStatusFilePath()); err != nil {
			log.Printf("failed to delete status file: %+v\n", err)
		}
	}()

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// StartRunner execute a setup script
func (s *AgentServer) StartRunner(ctx context.Context, req *pb.StartRunnerRequest) (*pb.StartRunnerResponse, error) {
	log.Println("call StartRunner")
	newSetupScript := appendSetStopStatus(req.SetupScript)

	go func() {
		// Execute background
		if err := execute(newSetupScript); err != nil {
			log.Printf("failed to execute setup script: %+v\n", err)
			return
		}
	}()

	return &pb.StartRunnerResponse{}, nil
}

func execute(setupScript string) error {
	f, err := os.OpenFile(getScriptFilePath(), os.O_WRONLY|os.O_CREATE, 0700)
	if err != nil {
		return fmt.Errorf("failed to open status file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(setupScript); err != nil {
		return fmt.Errorf("failed to write status file: %w", err)
	}

	if out, err := exec.Command(getScriptFilePath()).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to execute setup script: %+v out: %s", err, out)
	}

	return nil
}

func appendSetStopStatus(setupScript string) string {
	return fmt.Sprintf(`%s

## Set offline status for shoes-agent
echo -n %s > %s`, setupScript, pb.Status_Offline.String(), getStatusFilePath())
}
