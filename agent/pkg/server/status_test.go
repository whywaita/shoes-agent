//go:build linux || darwin

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	pb "github.com/whywaita/shoes-agent/proto.go"
)

func setup() *AgentServer {
	return &AgentServer{
		mu: sync.Mutex{},
	}
}

func TestAgentServer_setAgentStatus(t *testing.T) {
	as := setup()

	tests := []struct {
		input pb.Status
		want  pb.Status
		err   bool
	}{
		{
			input: pb.Status_Booting,
			want:  pb.Status_Booting,
			err:   false,
		},
	}

	for _, test := range tests {
		err := as.setAgentStatus(test.input)
		if err != nil && !test.err {
			t.Fatalf("setAgentStatus got error: %+v", err)
		}

		got, err := getAgentStatusFromFile()
		if err != nil {
			t.Fatalf("failed to get status from file: %+v\n", err)
		}
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}

func getAgentStatusFromFile() (pb.Status, error) {
	out, err := os.ReadFile(getFilePath())
	if err != nil {
		return -1, nil
	}

	return unmarshalStatus(string(out)), nil
}

func getFilePath() string {
	return filepath.Join(PathStatus, FileStatus)
}

// getAgentStatus get status of agent from status file
func (s *AgentServer) getAgentStatus() (pb.Status, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out, err := os.ReadFile(getFilePath())
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

	f, err := os.OpenFile(getFilePath(), os.O_WRONLY|os.O_CREATE, 0700)
	if err != nil {
		return fmt.Errorf("failed to open status file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprint(f, marshalStatus(status)); err != nil {
		return fmt.Errorf("failed to write status file: %w", err)
	}
	return nil
}
