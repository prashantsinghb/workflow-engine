package temporal

import (
	"time"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func WorkflowExecution(ctx workflow.Context, projectID string, g *dag.Graph, inputs map[string]interface{}) (map[dag.NodeID]map[string]interface{}, error) {
	done := map[dag.NodeID]bool{}
	outputs := map[dag.NodeID]map[string]interface{}{}

	for len(done) < len(g.Nodes) {
		progress := false

		for id, node := range g.Nodes {
			if done[id] {
				continue
			}

			ready := true
			for _, dep := range node.Depends {
				if !done[dep] {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}

			ao := workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
				RetryPolicy: &temporal.RetryPolicy{
					InitialInterval:    time.Second * 5,
					BackoffCoefficient: 2,
					MaximumInterval:    time.Minute,
					MaximumAttempts:    5,
				},
			}
			ctx1 := workflow.WithActivityOptions(ctx, ao)

			// Pass projectID to the activity
			var out map[string]interface{}
			err := workflow.ExecuteActivity(ctx1, NodeActivity, projectID, node, inputs).Get(ctx1, &out)
			if err != nil {
				return outputs, err
			}

			outputs[id] = out
			done[id] = true
			progress = true
		}

		if !progress {
			return outputs, temporal.NewApplicationError(
				"deadlock in DAG",
				"Deadlock",
				nil,
			)
		}
	}

	return outputs, nil
}
