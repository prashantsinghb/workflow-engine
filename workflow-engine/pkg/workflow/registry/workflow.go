package registry

import "github.com/prashantsinghb/workflow-engine/pkg/workflow/api"

type Workflow struct {
	ID        string
	ProjectID string
	Name      string
	Version   string
	Yaml      string
	Def       *api.Definition
}
