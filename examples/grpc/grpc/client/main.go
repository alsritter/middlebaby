package main

import (
	pb "github.com/alsritter/middlebaby/examples/helloworld/grpc/proto" // 引入proto包

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	// Address gRPC服务地址
	ADDRESS = "127.0.0.1:50052"
)

func main() {
	// 连接
	conn, err := grpc.Dial(ADDRESS, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalln(err)
	}
	defer conn.Close()

	// 初始化客户端
	c := pb.NewHelloClient(conn)

	// 调用方法
	req := &pb.HelloRequest{Name: "Hello gRPC !"}
	res, err := c.SayHello(context.Background(), req)

	if err != nil {
		grpclog.Fatalln(err)
	}

	grpclog.Info(res.Message)
}
