package main

import (
	"errors"
	"log"
	"math/rand"
	"os"

	"github.com/goyal-aman/mailmux/src/shared"
	"github.com/goyal-aman/mailmux/src/types"
	"github.com/hashicorp/go-plugin"
)

/*
Build command from project root
go build -o ./plugins/round_robin/round-robin-plugin ./plugins/round_robin/main.go

*/

// Here is a real implementation of Selector
type RoundRobinSelector struct{}

func (s *RoundRobinSelector) Select(downstreams []types.Downstream) (string, error) {
	log.Println("Executing RoundRobin plugin...")

	// Simple logic: just pick the first one for now, or implement actual round robin state
	// Note: Plugins are separate processes, so keeping state in memory works as long as the process lives.
	// However, the host might restart plugins.

	if len(downstreams) == 0 {
		return "", errors.New("no downstreams available")
	}

	N := len(downstreams)
	index := rand.Intn(N)
	// For demonstration, we just return the first one.
	// In a real round-robin, we'd need to persist state or use a random strategy if stateless.
	log.Println("Selected downstream:", downstreams[index].Addr)
	return downstreams[index].Addr, nil
}

func main() {
	// We don't want the plugin to log to stdout because that messes up the RPC protocol
	// So we can log to stderr
	log.SetOutput(os.Stderr)

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"selector": &shared.SelectorPlugin{Impl: &RoundRobinSelector{}},
		},
	})
}
