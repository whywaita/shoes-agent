package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
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
		return
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
		return fmt.Errorf("failed to execute setup script: %+v out: %s\n", err, out)
	}

	return nil
}

func appendSetStopStatus(setupScript string) string {
	return fmt.Sprintf(`%s

## Set offline status for shoes-agent
echo -n %s > %s`, setupScript, pb.Status_Offline.String(), getStatusFilePath())
}

var (
	errRunnerProcessIsNotFound   = fmt.Errorf("runner process is not found")
	errRunnerProcessIsNotRunning = fmt.Errorf("runner process is not running")
)

func (s *AgentServer) pollingRunnerProcess() error {
	log.Println("start polling runner process...")
	ctx := context.Background()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := searchRunnerProcess(ctx); err != nil {
				log.Printf("failed to search runner process, will retry...: %+v\n", err)
				continue
			}

			// found
			if err := s.setAgentStatus(pb.Status_Idle); err != nil {
				return fmt.Errorf("failed to set agent status (status: idle): %w", err)
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func searchRunnerProcess(ctx context.Context) error {
	log.Println("start searching process...")
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all process: %w", err)
	}

	for _, p := range processes {
		name, err := p.NameWithContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get name of process (process: %+v): %w", p, err)
		}

		if strings.EqualFold(name, "Runner.Listener") {
			if err := isRunningRunnerProcess(ctx, p); err != nil {
				return fmt.Errorf("failed to get running status: %w", err)
			}

			// found running runner!
			return nil
		}
	}

	return errRunnerProcessIsNotFound
}

func isRunningRunnerProcess(ctx context.Context, p *process.Process) error {
	cmd, err := p.CmdlineSliceWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cmdline: %w", err)
	}

	if strings.EqualFold(cmd[1], "run") {
		// found!
		log.Printf("found runner process! (pid: %d)\n", p.Pid)
		return nil
	}

	return fmt.Errorf("cmd: %s :%w", cmd, errRunnerProcessIsNotRunning)
}
