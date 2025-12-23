package executor

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
)

type DNSExecutor struct {
	pb.UnimplementedStepExecutorServer
}

func (d *DNSExecutor) ExecuteStep(
	ctx context.Context,
	req *pb.ExecuteRequest,
) (*pb.ExecuteResponse, error) {

	domain := req.Input["var1"]
	fmt.Println("domain", domain)
	return &pb.ExecuteResponse{
		Output: map[string]string{
			"var1": "var1-" + domain,
		},
	}, nil
}

func RegisterStep() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
	client := pb.NewStepRegistryClient(conn)

	_, err := client.RegisterSteps(context.Background(), &pb.RegisterStepsRequest{
		Service: "test-service",
		Steps: []*pb.StepSpec{
			{
				Name:     "test.print",
				Version:  "v1",
				Protocol: "grpc",
				Endpoint: "localhost:7078",
				InputSchema: map[string]string{
					"var1": "string",
				},
				OutputSchema: map[string]string{
					"output_name": "string",
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}
