package main

import (
	context "context"
	"log"
	"net"

	pb "github.com/alsritter/middlebaby/examples/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	ADDRESS = "127.0.0.1:50052"
)

var _ pb.TestServiceServer = (*testService)(nil)

type testService struct{}

// Create implements proto.TestServiceServer
func (*testService) Create(context.Context, *pb.CreateRequest) (*pb.CreateResponse, error) {
	panic("unimplemented")
}

// GetById implements proto.TestServiceServer
func (*testService) GetById(context.Context, *pb.GetByIdRequest) (*pb.GetByIdResponse, error) {
	panic("unimplemented")
}

// GetList implements proto.TestServiceServer
func (*testService) GetList(context.Context, *pb.GetListRequest) (*pb.GetListResponse, error) {
	panic("unimplemented")
}

// Update implements proto.TestServiceServer
func (*testService) Update(context.Context, *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	panic("unimplemented")
}

func main() {
	listen, err := net.Listen("tcp", ADDRESS)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 实例化 grpc Server
	s := grpc.NewServer()
	pb.RegisterTestServiceServer(s, new(testService))
	reflection.Register(s)
	log.Println("Listen on " + ADDRESS)
	s.Serve(listen)
}
