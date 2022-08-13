package main

import (
	"context"
	"fmt"
	"net"

	pb "github.com/alsritter/middlebaby/examples/helloworld/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

const (
	// Address gRPC服务地址
	ADDRESS = "127.0.0.1:50052"
)

// 定义 HelloService 并实现约定的接口
type HelloService struct{}

// SayHello 实现 Hello 服务接口
func (h *HelloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("Hello %s.", in.Name)
	return resp, nil
}

func main() {
	listen, err := net.Listen("tcp", ADDRESS)
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	// HelloService Hello 服务
	helloService := new(HelloService)

	// 实例化 grpc Server
	s := grpc.NewServer()

	// 注册 HelloService
	pb.RegisterHelloServer(s, helloService)

	reflection.Register(s)

	grpclog.Infoln("Listen on " + ADDRESS)
	s.Serve(listen)
}
