package server

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_appendSetStopStatus(t *testing.T) {
	fp := getStatusFilePath()

	tests := []struct {
		input string
		want  string
		err   bool
	}{

		{
			input: `#!/bin/bash

echo 000`,
			want: fmt.Sprintf(`#!/bin/bash

echo 000

## Set offline status for shoes-agent
echo -n Offline > %s`, fp),
			err: false,
		},
	}

	for _, test := range tests {
		got := appendSetStopStatus(test.input)
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}
