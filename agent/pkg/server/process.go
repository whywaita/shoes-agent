package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
	pb "github.com/whywaita/shoes-agent/proto.go"
)

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
