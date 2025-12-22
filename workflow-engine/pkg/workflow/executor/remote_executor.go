package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"

	mRegistry "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"google.golang.org/grpc"
)

// RemoteExecutor resolves steps via ModuleRegistry and executes them over HTTP/gRPC
type RemoteExecutor struct {
	modules *mRegistry.ModuleRegistry
}

func NewRemoteExecutor(modules *mRegistry.ModuleRegistry) *RemoteExecutor {
	return &RemoteExecutor{modules: modules}
}

// Execute executes a remote step
func (r *RemoteExecutor) Execute(ctx context.Context, node *dag.Node, inputs map[string]interface{}) (map[string]interface{}, error) {
	mod, err := r.modules.GetModule(ctx, "", node.Uses, "")
	if err != nil {
		return nil, errors.New("step not found in registry: " + node.Executor)
	}

	protocol, ok := mod.RuntimeConfig["protocol"].(string)
	if !ok {
		return nil, errors.New("protocol not found or invalid in runtime config")
	}

	endpoint, ok := mod.RuntimeConfig["endpoint"].(string)
	if !ok {
		return nil, errors.New("endpoint not found or invalid in runtime config")
	}

	if protocol == "http" {
		return executeHTTP(endpoint, node.Executor, inputs)
	} else if protocol == "grpc" {
		return executeGRPC(endpoint, node.Executor, inputs)
	}
	return nil, errors.New("unsupported protocol: " + protocol)
}

// -------------------- HTTP execution --------------------

func executeHTTP(endpoint, step string, input map[string]interface{}) (map[string]interface{}, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"step":  step,
		"input": input,
	})
	resp, err := http.Post(endpoint+"/workflow/execute", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Output map[string]interface{} `json:"output"`
		Error  string                 `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	return result.Output, nil
}

// -------------------- gRPC execution --------------------

func executeGRPC(endpoint, step string, input map[string]interface{}) (map[string]interface{}, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewStepExecutorClient(conn)
	req := &pb.ExecuteRequest{
		Step:  step,
		Input: mapToStringMap(input),
	}

	resp, err := client.ExecuteStep(context.Background(), req)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	output := make(map[string]interface{})
	for k, v := range resp.Output {
		output[k] = v
	}
	return output, nil
}

// Helper: convert interface{} -> string map for gRPC
func mapToStringMap(input map[string]interface{}) map[string]string {
	m := make(map[string]string)
	for k, v := range input {
		m[k] = fmt.Sprintf("%v", v)
	}
	return m
}
