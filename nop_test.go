package events

import (
	"errors"
	"io"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/types"
)

var (
	errNoEngine = errors.New("Engine no longer exists")
)

// NopClient is a nop API Client based on engine-api
type NopClient struct {
}

// NewNopClient creates a new nop client
func NewNopClient() *NopClient {
	return &NopClient{}
}

// ClientVersion returns the version string associated with this instance of the Client
func (client *NopClient) ClientVersion() string {
	return ""
}

// Events returns a stream of events in the daemon in a ReadCloser
func (client *NopClient) Events(ctx context.Context, options types.EventsOptions) (io.ReadCloser, error) {
	return nil, errNoEngine
}

// Info returns information about the docker server
func (client *NopClient) Info(ctx context.Context) (types.Info, error) {
	return types.Info{}, errNoEngine
}

// RegistryLogin authenticates the docker server with a given docker registry
func (client *NopClient) RegistryLogin(ctx context.Context, auth types.AuthConfig) (types.AuthResponse, error) {
	return types.AuthResponse{}, errNoEngine
}
