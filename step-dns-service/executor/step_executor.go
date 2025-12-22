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

	domain := req.Input["domain"]
	fmt.Println("domain", domain)
	return &pb.ExecuteResponse{
		Output: map[string]string{
			"record_id": "dns-" + domain,
		},
	}, nil
}

func RegisterStep() {
	conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
	client := pb.NewStepRegistryClient(conn)

	_, err := client.RegisterSteps(context.Background(), &pb.RegisterStepsRequest{
		Service: "dns-service",
		Steps: []*pb.StepSpec{
			{
				Name:     "dns.create",
				Version:  "v4",
				Protocol: "grpc",
				Endpoint: "localhost:7070",
				InputSchema: map[string]string{
					"domain": "string",
				},
				OutputSchema: map[string]string{
					"record_id": "string",
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}
}
