/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package grpchandler

import (
	"context"
	"fmt"
	"net/textproto"
	"path/filepath"
	"strings"

	"github.com/alsritter/middlebaby/pkg/messagepush"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/types/interact"
	"github.com/alsritter/middlebaby/pkg/util/grpcurl/ext/ggrpcurl"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
	"github.com/alsritter/middlebaby/pkg/util/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/hashicorp/go-multierror"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type captureServer struct {
	descriptor proto.Descriptor
	logger.Logger
	protoManager protomanager.Provider
}

type Provider interface {
	Init(ctx *mbcontext.Context) error
	GetServer() *grpc.Server
}

func New(log logger.Logger, protoManager protomanager.Provider) Provider {
	return &captureServer{
		Logger:       log.NewLogger("grpc-capture"),
		protoManager: protoManager,
	}
}

func (s *captureServer) Init(ctx *mbcontext.Context) error {
	s.Info(nil, "stating proto manager")
	if err := s.protoManager.Start(ctx); err != nil {
		return err
	}
	return nil
}

func (s *captureServer) GetServer() *grpc.Server {
	return grpc.NewServer(grpc.UnknownServiceHandler(s.handleStream))
}

func (s *captureServer) handleStream(srv interface{}, stream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return s.sendError(stream.Context(), status.Errorf(codes.Internal, "lowLevelServerStream not exists in context"))
	}
	md, _ := metadata.FromIncomingContext(stream.Context())

	s.WithContext(stream.Context()).Debug(map[string]interface{}{"path": fullMethodName, "metadata": md}, "request received")

	method, ok := s.protoManager.GetMethod(fullMethodName)
	if !ok {
		return s.sendError(stream.Context(), status.Errorf(codes.NotFound, "method not found"))
	}
	request := dynamic.NewMessage(method.GetInputType())
	// receive request
	if err := stream.RecvMsg(request); err != nil {
		return s.sendError(stream.Context(), status.Errorf(codes.Unknown, "failed to recv request"))
	}
	data, err := request.MarshalJSONPB(&jsonpb.Marshaler{})
	if err != nil {
		return s.sendError(stream.Context(), status.Errorf(codes.Unknown, "failed to marshal request"))
	}

	serviceMethod := strings.TrimPrefix(fullMethodName, "/")
	dto := ggrpcurl.GGrpCurlDTO{
		Plaintext:     true,
		FormatError:   true,
		EmitDefaults:  true,
		AddHeaders:    copyToGGrpCurlHeader(md),
		ImportPaths:   s.protoManager.GetImportPaths(),
		ProtoFiles:    []string{filepath.Base(method.GetService().GetFile().GetName())},
		Data:          string(data),
		ServiceAddr:   getAuthorityFromMetadata(md),
		ServiceMethod: serviceMethod,
	}

	if err != nil {
		return s.sendError(stream.Context(), err)
	}

	messagepush.SendMessage(dto)

	s.WithContext(stream.Context()).Debug(nil, "capture [%s] request [%+v]", fullMethodName, dto)
	mds, trailer, responseStr, respStatus, err := ggrpcurl.NewInvokeGRpc(&dto).Invoke()
	if err != nil {
		return s.sendError(stream.Context(), err)
	}

	s.WithContext(stream.Context()).Debug(nil, "capture [%s], response [%s], headers [%v], trailer [%v]", fullMethodName, responseStr, mds, trailer)

	stream.SetTrailer(trailer)
	if len(mds) > 0 {
		if err := stream.SetHeader(getMetadataFromHeaderMap(mds)); err != nil {
			return s.sendError(stream.Context(), status.Errorf(codes.Unavailable, "failed to set header: %s", err))
		}
	}

	if respStatus.Code() != 0 {
		return s.sendError(stream.Context(), status.Errorf(respStatus.Code(), "expected code is: %d", respStatus.Code()))
	}

	if !ok {
		return s.sendError(stream.Context(), fmt.Errorf("unable to find descriptor: %s", fullMethodName))
	}

	messagepush.SendMessage(responseStr)

	message := dynamic.NewMessage(method.GetOutputType())
	if err := message.UnmarshalJSONPB(&jsonpb.Unmarshaler{}, []byte(responseStr)); err != nil {
		return s.sendError(stream.Context(), multierror.Prefix(err, "failed to unmarshal:"))
	}

	binaryData, err := message.Marshal()
	if err != nil {
		return s.sendError(stream.Context(), multierror.Prefix(err, "failed to marshal:"))
	}

	// send the response
	if err := stream.SendMsg(interact.NewBytesMessage(binaryData)); err != nil {
		return s.sendError(stream.Context(), status.Errorf(codes.Internal, "failed to send message: %s", err))
	}
	return nil
}

func copyToGGrpCurlHeader(h map[string][]string) (headers []string) {
	for k := range h {
		headers = append(headers, fmt.Sprintf("%s:%s", k, textproto.MIMEHeader(h).Get(k)))
	}
	return headers
}

func (s *captureServer) sendError(ctx context.Context, err error) error {
	s.WithContext(ctx).Error(nil, "%v", err)
	return err
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
