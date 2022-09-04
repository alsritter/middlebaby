package main

import (
	context "context"
	"fmt"
	"log"
	"net"

	pb "github.com/alsritter/middlebaby/examples/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

const (
	ADDRESS = ":50052"
)

var _ pb.TestServiceServer = (*testService)(nil)

type testService struct{}

// Create implements proto.TestServiceServer
func (*testService) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	cResp, _ := GetOutsideClient().PutData(ctx, &pb.PutRequest{
		Name:   req.Info.Name,
		Age:    req.Info.Age,
		Gender: req.Info.Gender,
	})

	fmt.Printf("=========== %v \n", cResp)

	return &pb.CreateResponse{
		ActivityId: fmt.Sprintf("%s-%t", req.ProjectId, cResp.Status),
	}, nil
}

// GetById implements proto.TestServiceServer
func (*testService) GetById(ctx context.Context, req *pb.GetByIdRequest) (*pb.GetByIdResponse, error) {
	panic("unimplemented")
}

// GetList implements proto.TestServiceServer
func (*testService) GetList(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	panic("unimplemented")
}

// Update implements proto.TestServiceServer
func (*testService) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	panic("unimplemented")
}

func main() {
	// 随便一个地址，反正这里会被 Mock
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

var outsideClient pb.OutsideServiceClient

func GetOutsideClient() pb.OutsideServiceClient {
	if outsideClient == nil {
		conn, err := grpc.Dial(":56789", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			grpclog.Fatalln(err)
		}
		outsideClient := pb.NewOutsideServiceClient(conn)
		return outsideClient
	}

	return outsideClient
}
