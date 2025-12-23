package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
)

// UserExecutor implements the StepExecutor gRPC server
type UserExecutor struct {
	pb.UnimplementedStepExecutorServer

	// Map step names to handler functions
	handlers map[string]func(context.Context, map[string]string) (map[string]string, error)
}

// ExecuteStep routes execution to the correct handler based on the step name
func (u *UserExecutor) ExecuteStep(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	handler, ok := u.handlers[req.Step]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "handler not found for step: %s", req.Step)
	}

	out, err := handler(ctx, req.Input)
	if err != nil {
		return nil, err
	}

	return &pb.ExecuteResponse{Output: out}, nil
}

// ExecuteRolledBackStep routes rollback execution to the correct handler
func (u *UserExecutor) ExecuteRolledBackStep(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	handlerName := req.Step + ".rollback"
	handler, ok := u.handlers[handlerName]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "rollback handler not found for step: %s", req.Step)
	}

	out, err := handler(ctx, req.Input)
	if err != nil {
		return nil, err
	}

	return &pb.ExecuteResponse{Output: out}, nil
}

// registerSteps registers all steps with the Step Registry service
func registerSteps() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pb.NewStepRegistryClient(conn)

	steps := []*pb.StepSpec{
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
		{
			Name:     "user.create",
			Version:  "v1",
			Protocol: "grpc",
			Endpoint: "localhost:8082",
			InputSchema: map[string]string{
				"userId": "string",
				"name":   "string",
			},
			OutputSchema: map[string]string{
				"userId": "string",
				"status": "string",
			},
		},
		{
			Name:     "user.delete",
			Version:  "v1",
			Protocol: "grpc",
			Endpoint: "localhost:8082",
			InputSchema: map[string]string{
				"userId": "string",
			},
			OutputSchema: map[string]string{
				"status": "string",
			},
		},
	}

	_, err = client.RegisterSteps(context.Background(), &pb.RegisterStepsRequest{
		Service: "user-service",
		Steps:   steps,
	})
	if err != nil {
		panic(err)
	}

	log.Println("Steps registered successfully!")
}

func main() {
	// Initialize executor and handlers
	executor := &UserExecutor{
		handlers: map[string]func(context.Context, map[string]string) (map[string]string, error){
			"user.fetch": func(ctx context.Context, input map[string]string) (map[string]string, error) {
				log.Println("Executing user.fetch")
				userID := input["userId"]
				return map[string]string{
					"userId": userID,
					"email":  userID + "@example.com",
				}, nil
			},
			"user.create": func(ctx context.Context, input map[string]string) (map[string]string, error) {
				log.Println("Executing user.create")
				userID := input["userId"]
				name := input["name"]
				return map[string]string{
					"userId": userID,
					"status": "Created user " + name,
				}, nil
			},
			"user.delete": func(ctx context.Context, input map[string]string) (map[string]string, error) {
				log.Println("Executing user.delete")
				userID := input["userId"]
				return map[string]string{
					"status": "Deleted user " + userID,
				}, nil
			},
			// Optional rollback handlers
			"user.create.rollback": func(ctx context.Context, input map[string]string) (map[string]string, error) {
				log.Println("Executing user.create.rollback")
				userID := input["userId"]
				return map[string]string{
					"message": "Rolled back creation of " + userID,
				}, nil
			},
		},
	}

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStepExecutorServer(grpcServer, executor)

	// Register steps in a separate goroutine
	go registerSteps()

	log.Println("User service listening on :8082")
	log.Fatal(grpcServer.Serve(lis))
}
