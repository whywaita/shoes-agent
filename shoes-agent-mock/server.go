package main

import (
	"log"

	"github.com/hashicorp/go-plugin"
	"github.com/whywaita/shoes-agent/shoes-agent-mock/mock"
	"github.com/whywaita/shoes-agent/shoes-agent/pkg/shoesprovider"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	m := mock.New()

	handshake := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "SHOES_PLUGIN_MAGIC_COOKIE",
		MagicCookieValue: "are_you_a_shoes?",
	}
	pluginMap := map[string]plugin.Plugin{
		"shoes_grpc": &shoesprovider.AgentPlugin{
			Backend:   m,
			ShoesType: "shoes-agent-mock",
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})

	return nil
}
