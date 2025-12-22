package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
)

type UserExecutor struct {
	pb.UnimplementedStepExecutorServer
}

func (u *UserExecutor) ExecuteStep(
	ctx context.Context,
	req *pb.ExecuteRequest,
) (*pb.ExecuteResponse, error) {
	userID := req.Input["userId"]
	log.Println("User service executed:", userID)
	out := map[string]string{
		"userId": userID,
		"email":  userID + "@example.com",
	}

	log.Println("User service executed:", out)

	return &pb.ExecuteResponse{Output: out}, nil
}

func registerStep() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
	client := pb.NewStepRegistryClient(conn)

	_, err := client.RegisterSteps(context.Background(), &pb.RegisterStepsRequest{
		Service: "user-service",
		Steps: []*pb.StepSpec{
			{
				Name:     "user.fetch",
				Version:  "v1",
				Protocol: "grpc",
				Endpoint: "localhost:8082",
				InputSchema: map[string]string{
					"userId": "string",
				},
				OutputSchema: map[string]string{
					"userId": "string",
					"email":  "string",
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}

func main() {
	lis, _ := net.Listen("tcp", ":8082")
	grpcServer := grpc.NewServer()

	pb.RegisterStepExecutorServer(grpcServer, &UserExecutor{})

	go registerStep()

	log.Println("User service listening on :8082")
	log.Fatal(grpcServer.Serve(lis))
}
