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
type Config struct {
	Address      string
	ProtoManager *protomanager.Config
}

type mockServer struct {
	cfg *Config
	logger.Logger
	apiManager   apimanager.Provider
	protoManager protomanager.Provider
}

type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) (*grpc.Server, error)
}

func New(log logger.Logger, cfg *Config, apiManager apimanager.Provider, protoManager protomanager.Provider) (Provider, error) {
	m := &mockServer{
		cfg:          cfg,
		Logger:       log.NewLogger("grpcMockServer"),
		apiManager:   apiManager,
		protoManager: protoManager,
	}
	if err := m.setup(); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *mockServer) Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) (*grpc.Server, error) {
	s.Info(nil, "stating proto manager")
	if err := s.protoManager.Start(ctx, cancelFunc, wg); err != nil {
		return nil, err
	}

	return grpc.NewServer(grpc.UnknownServiceHandler(s.handleStream)), nil
}

func (s *mockServer) setup() error {
	if err := s.setupProtoManager(); err != nil {
		return err
	}
	return nil
}

func (s *mockServer) setupProtoManager() error {
	service, err := protomanager.New(s.cfg.ProtoManager, s.Logger)
	if err != nil {
		return err
	}
	s.protoManager = service
	return nil
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
		Headers:  getHeadersFromMetadata(md),
		Body:     interact.NewBytesMessage(data),
	})
	if err != nil {
		return err
	}
	stream.SetTrailer(metadata.New(response.Trailer))
	if len(response.Headers) > 0 {
		if err := stream.SetHeader(metadata.New(response.Headers)); err != nil {
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
