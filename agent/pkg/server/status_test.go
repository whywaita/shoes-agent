package server

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	pb "github.com/whywaita/shoes-agent/proto.go"
)

func TestAgentServer_setAgentStatus(t *testing.T) {
	as := New()

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
	out, err := os.ReadFile(getStatusFilePath())
	if err != nil {
		return -1, nil
	}

	return unmarshalStatus(string(out)), nil
}
