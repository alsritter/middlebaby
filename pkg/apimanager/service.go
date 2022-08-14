package apimanager

import (
	"context"
	"net/http"

	"github.com/alsritter/middlebaby/pkg/interact"
)

type Provider interface {
	MockResponse(ctx context.Context, request *http.Request) (*interact.GRpcResponse, error)
}

type Manager struct {
	MockCenter
}
