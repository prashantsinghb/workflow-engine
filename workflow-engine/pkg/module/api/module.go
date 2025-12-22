package api

import (
	"time"
)

type Module struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	ProjectID     string                 `json:"project_id,omitempty"` // nil for global
	Runtime       string                 `json:"runtime"`              // http/docker
	RuntimeConfig map[string]interface{} `json:"runtime_config,omitempty"`
	Inputs        map[string]interface{} `json:"inputs,omitempty"`
	Outputs       map[string]interface{} `json:"outputs,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}
