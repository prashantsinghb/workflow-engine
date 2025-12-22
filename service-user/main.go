package main

import (
	"context"
	"log"
	"net"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
)

type UserExecutor struct {
	pb.UnimplementedStepExecutorServer
	// counter to simulate transient failures
	failCounter int32
}

// ExecuteStep simulates retries: fails first 2 attempts, succeeds on 3rd
func (u *UserExecutor) ExecuteStep(
	ctx context.Context,
	req *pb.ExecuteRequest,
) (*pb.ExecuteResponse, error) {
	attempt := atomic.AddInt32(&u.failCounter, 1)
	userID := req.Input["userId"]

	if attempt <= 2 {
		log.Printf("Simulated failure for user %s, attempt %d\n", userID, attempt)
		return nil, status.Error(codes.Unavailable, "transient error, please retry")
	}

	out := map[string]string{
		"userId": userID,
		"email":  userID + "@example.com",
	}

	log.Printf("User service executed successfully on attempt %d: %+v\n", attempt, out)
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
