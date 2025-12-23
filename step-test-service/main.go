package main

import (
	"net"
	"time"

	"github.com/prashantsinghb/step-dns-service/executor"
	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
	"google.golang.org/grpc"
)

func main() {
	go startExecutor()

	time.Sleep(1 * time.Second) // ensure executor up

	executor.RegisterStep()

	select {} // keep running
}

func startExecutor() {
	lis, _ := net.Listen("tcp", ":7074")
	grpcServer := grpc.NewServer()

	pb.RegisterStepExecutorServer(grpcServer, &executor.DNSExecutor{})
	grpcServer.Serve(lis)
}
