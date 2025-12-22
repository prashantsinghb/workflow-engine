package server

import (
	"context"
	"encoding/json"
	"net/http"

	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
	"google.golang.org/grpc"
)

type HTTPRegisterStepsRequest struct {
	Service string        `json:"service"`
	Steps   []StepSpecDTO `json:"steps"`
}

type StepSpecDTO struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Protocol     string            `json:"protocol"`
	Endpoint     string            `json:"endpoint"`
	InputSchema  map[string]string `json:"input_schema"`
	OutputSchema map[string]string `json:"output_schema"`
}

// HTTP wrapper: converts HTTP request to gRPC call
func HTTPRegisterStepsHandler(grpcAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req HTTPRegisterStepsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Dial gRPC engine
		conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		client := pb.NewStepRegistryClient(conn)

		grpcReq := &pb.RegisterStepsRequest{
			Service: req.Service,
			Steps:   []*pb.StepSpec{},
		}

		for _, s := range req.Steps {
			grpcReq.Steps = append(grpcReq.Steps, &pb.StepSpec{
				Name:         s.Name,
				Version:      s.Version,
				Protocol:     s.Protocol,
				Endpoint:     s.Endpoint,
				InputSchema:  s.InputSchema,
				OutputSchema: s.OutputSchema,
			})
		}

		resp, err := client.RegisterSteps(context.Background(), grpcReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(resp)
	}
}
