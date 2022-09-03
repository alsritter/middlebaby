package grpchandler

import (
	"context"
	"net/http"
	"sync"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Config defines the config structure
type Config struct{}

type mockServer struct {
	logger.Logger
	apiManager   apimanager.Provider
	protoManager protomanager.Provider
}

type Provider interface {
	Init(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error
	GetServer() *grpc.Server
}

func New(log logger.Logger, apiManager apimanager.Provider, protoManager protomanager.Provider) Provider {
	return &mockServer{
		Logger:       log.NewLogger("grpc"),
		apiManager:   apiManager,
		protoManager: protoManager,
	}
}

func (s *mockServer) Init(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error {
	s.Info(nil, "stating proto manager")
	if err := s.protoManager.Start(ctx, cancelFunc, wg); err != nil {
		return err
	}
	return nil
}

func (s *mockServer) GetServer() *grpc.Server {
	return grpc.NewServer(grpc.UnknownServiceHandler(s.handleStream))
}

func (s *mockServer) handleStream(srv interface{}, stream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}

	md, _ := metadata.FromIncomingContext(stream.Context())
	s.Info(map[string]interface{}{"path": fullMethodName, "metadata": md}, "request received")

	method, ok := s.protoManager.GetMethod(fullMethodName)
	if !ok {
		return status.Errorf(codes.NotFound, "method not found")
	}
	request := dynamic.NewMessage(method.GetInputType())
	// receive request
	if err := stream.RecvMsg(request); err != nil {
		return status.Errorf(codes.Unknown, "failed to recv request")
	}

	data, err := request.MarshalJSONPB(&jsonpb.Marshaler{})
	if err != nil {
		return status.Errorf(codes.Unknown, "failed to marshal request")
	}
	response, err := s.apiManager.MockResponse(context.TODO(), &interact.Request{
		Protocol: interact.ProtocolGRPC,
		Method:   http.MethodPost,
		Host:     getAuthorityFromMetadata(md),
		Path:     fullMethodName,
		Header:   getHeadersFromMetadata(md),
		Body:     interact.NewBytesMessage(data),
	})
	if err != nil {
		return err
	}
	stream.SetTrailer(metadata.New(response.Trailer))
	if len(response.Header) > 0 {
		if err := stream.SetHeader(getMetadataFromHeaderMap(response.Header)); err != nil {
			return status.Errorf(codes.Unavailable, "failed to set header: %s", err)
		}
	}
	if response.Status != 0 {
		return status.Errorf(codes.Code(response.Status), "expected code is: %d", response.Status)
	}

	// send the response
	if err := stream.SendMsg(response.Body); err != nil {
		return status.Errorf(codes.Internal, "failed to send message: %s", err)
	}
	return nil
}

// getHeadersFromMetadata is used to convert Metadata to Headers
func getHeadersFromMetadata(md metadata.MD) map[string]interface{} {
	headers := map[string]interface{}{}
	for key, values := range md {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// getAuthorityFromMetadata is used to get authority from metadata
func getAuthorityFromMetadata(md metadata.MD) string {
	if md != nil {
		values := md[":authority"]
		if len(values) != 0 {
			return values[0]
		}
	}
	return ""
}

// func getHeadersFromMetadata
func getMetadataFromHeaderMap(headers map[string][]string) metadata.MD {
	tmp := make(map[string]string)
	for k, v := range headers {
		tmp[k] = v[0]
	}
	return metadata.New(tmp)
}
