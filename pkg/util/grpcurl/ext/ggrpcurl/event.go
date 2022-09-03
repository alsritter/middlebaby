package ggrpcurl

import (
	"github.com/alsritter/middlebaby/pkg/util/grpcurl"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CustomEventHandler struct {
	*grpcurl.DefaultEventHandler
	ResponseMd metadata.MD
	TrailersMd metadata.MD
}

func (h *CustomEventHandler) OnReceiveHeaders(md metadata.MD) {
	h.DefaultEventHandler.OnReceiveHeaders(md)
	h.ResponseMd = md
}

func (h *CustomEventHandler) OnReceiveTrailers(stat *status.Status, md metadata.MD) {
	h.DefaultEventHandler.OnReceiveTrailers(stat, md)
	h.TrailersMd = md
}
