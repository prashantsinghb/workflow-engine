package parser

import (
	"fmt"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
	"sigs.k8s.io/yaml"
)

func ParseWorkflow(data []byte) (*api.Definition, error) {
	var def api.Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow definition: %w", err)
	}
	if len(def.Nodes) == 0 {
		return nil, workflow.ErrEmptyWorkflow
	}
	return &def, nil
}
