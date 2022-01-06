package main

import (
	"fmt"
	"log"

	"github.com/whywaita/shoes-agent/agent/pkg/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	s := server.New()
	if err := s.Run(); err != nil {
		return fmt.Errorf("failed to run gRPC server: %w", err)
	}

	return nil
}
