package workflow

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

func Parse(data []byte) (*Definition, error) {
	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow definition: %w", err)
	}
	if len(def.Nodes) == 0 {
		return nil, ErrEmptyWorkflow
	}
	return &def, nil
}
