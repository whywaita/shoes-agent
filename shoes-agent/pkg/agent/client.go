package agent

import (
	"fmt"

	"google.golang.org/grpc"
)

type Client struct {
	Conn *grpc.ClientConn
}

func NewClient(agentEndpoint string) (*Client, error) {
	grpcConn, err := grpc.Dial(
		agentEndpoint,
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial agent endpoint (endpoint: %s): %w", agentEndpoint, err)
	}

	return &Client{
		Conn: grpcConn,
	}, nil
}
