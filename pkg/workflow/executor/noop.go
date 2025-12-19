package executor

import (
	"fmt"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
)

func RunNode(node api.Node, inputs map[string]interface{}) (map[string]interface{}, error) {
	switch node.Uses {
	case "noop":
		return inputs, nil
	default:
		return nil, fmt.Errorf("unknown executor type: %s", node.Uses)
	}
}
