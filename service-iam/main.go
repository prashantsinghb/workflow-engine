package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
)

type IAMExecutor struct {
	pb.UnimplementedStepExecutorServer
}

func (i *IAMExecutor) ExecuteStep(
	ctx context.Context,
	req *pb.ExecuteRequest,
) (*pb.ExecuteResponse, error) {

	email := req.Input["email"]

	out := map[string]string{
		"assigned": "true",
		"email":    email,
		"role":     "admin",
	}

	log.Println("IAM service executed:", out)

	return &pb.ExecuteResponse{Output: out}, nil
}

func registerStep() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
	client := pb.NewStepRegistryClient(conn)

	_, err := client.RegisterSteps(context.Background(), &pb.RegisterStepsRequest{
		Service: "iam-service",
		Steps: []*pb.StepSpec{
			{
				Name:     "iam.assignRole",
				Version:  "v1",
				Protocol: "grpc",
				Endpoint: "localhost:8083",
				InputSchema: map[string]string{
					"email": "string",
				},
				OutputSchema: map[string]string{
					"assigned": "bool",
					"email":    "string",
					"role":     "string",
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}

func main() {
	lis, _ := net.Listen("tcp", ":8083")
	grpcServer := grpc.NewServer()

	pb.RegisterStepExecutorServer(grpcServer, &IAMExecutor{})

	go registerStep()

	log.Println("IAM service listening on :8083")
	log.Fatal(grpcServer.Serve(lis))
}
