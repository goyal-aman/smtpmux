package shared

import (
	"net/rpc"

	"github.com/goyal-aman/mailmux/src/types"
	"github.com/hashicorp/go-plugin"
)

// Selector is the interface that we're exposing as a plugin.
type Selector interface {
	Select(downstreams []types.Downstream) (string, error)
}

// Here is an implementation that talks over RPC
type SelectorRPC struct{ client *rpc.Client }

func (g *SelectorRPC) Select(downstreams []types.Downstream) (string, error) {
	var resp string
	err := g.client.Call("Plugin.Select", downstreams, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

// Here is the RPC server that the plugin runs
type SelectorRPCServer struct {
	Impl Selector
}

func (s *SelectorRPCServer) Select(args []types.Downstream, resp *string) error {
	var err error
	*resp, err = s.Impl.Select(args)
	return err
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a SelectorRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return SelectorRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type SelectorPlugin struct {
	// Impl Injection
	Impl Selector
}

func (p *SelectorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &SelectorRPCServer{Impl: p.Impl}, nil
}

func (p *SelectorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &SelectorRPC{client: c}, nil
}

// HandshakeConfig is a shared handshake config that the client and server must agree on
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SMTP_ROUTER_PLUGIN",
	MagicCookieValue: "hello",
}
