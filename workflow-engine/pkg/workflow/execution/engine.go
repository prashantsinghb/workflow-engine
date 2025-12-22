package execution

import (
	"context"
	"fmt"
	"sync"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
	wfRegistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

type State string

const (
	StateRunning State = "RUNNING"
	StateSuccess State = "SUCCESS"
	StateFailed  State = "FAILED"
)

type Execution struct {
	ID       string
	Project  string
	Workflow *wfRegistry.Workflow
	State    State
	Outputs  map[dag.NodeID]map[string]interface{}
	Error    string
}

var (
	mu         sync.Mutex
	executions = map[string]map[string]*Execution{}
	counter    int
)

func Start(
	ctx context.Context,
	projectID string,
	workflowID string,
	inputs map[string]interface{},
	executionContext *Context,
) (string, error) {

	ctx = executor.WithProjectID(ctx, projectID)

	wf, err := executionContext.Workflow.Get(ctx, projectID, workflowID)
	if err != nil {
		return "", err
	}

	graph := dag.Build(wf.Def)

	mu.Lock()
	counter++
	execID := fmt.Sprintf("%s-exec-%d", workflowID, counter)

	exec := &Execution{
		ID:       execID,
		Project:  projectID,
		Workflow: wf,
		State:    StateRunning,
		Outputs:  make(map[dag.NodeID]map[string]interface{}),
	}

	if executions[projectID] == nil {
		executions[projectID] = make(map[string]*Execution)
	}
	executions[projectID][execID] = exec
	mu.Unlock()

	done := make(map[dag.NodeID]bool)

	for len(done) < len(graph.Nodes) {
		progress := false

		for id, node := range graph.Nodes {
			if done[id] {
				continue
			}

			if !isReady(node, done) {
				continue
			}

			mod, err := executionContext.Modules.Resolve(ctx, projectID, node.Uses)
			if err != nil {
				exec.State = StateFailed
				exec.Error = err.Error()
				return execID, nil
			}

			execImpl, err := executor.Get(mod.Runtime)
			if err != nil {
				exec.State = StateFailed
				exec.Error = err.Error()
				return execID, nil
			}

			out, err := execImpl.Execute(ctx, node, inputs)
			if err != nil {
				exec.State = StateFailed
				exec.Error = err.Error()
				return execID, nil
			}

			exec.Outputs[id] = out
			done[id] = true
			progress = true
		}

		if !progress {
			exec.State = StateFailed
			exec.Error = "deadlock detected in DAG execution"
			return execID, nil
		}
	}

	exec.State = StateSuccess
	return execID, nil
}

func GetExecution(projectID, execID string) (*Execution, error) {
	mu.Lock()
	defer mu.Unlock()

	if proj, ok := executions[projectID]; ok {
		if exec, ok := proj[execID]; ok {
			return exec, nil
		}
	}
	return nil, fmt.Errorf("execution not found")
}

func isReady(node *dag.Node, done map[dag.NodeID]bool) bool {
	for _, dep := range node.Depends {
		if !done[dep] {
			return false
		}
	}
	return true
}
