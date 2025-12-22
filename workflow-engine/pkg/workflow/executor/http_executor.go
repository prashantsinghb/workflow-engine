package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	wfRegistry "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/dag"
)

// HttpExecutor executes modules via HTTP
type HttpExecutor struct {
	modules *wfRegistry.ModuleRegistry
	client  *http.Client
}

// NewHttpExecutor creates a new HTTP executor
func NewHttpExecutor(modules *wfRegistry.ModuleRegistry) *HttpExecutor {
	return &HttpExecutor{
		modules: modules,
		client:  &http.Client{},
	}
}

// Execute performs the HTTP call based on module spec
func (e *HttpExecutor) Execute(
	ctx context.Context,
	node *dag.Node,
	inputs map[string]interface{},
) (map[string]interface{}, error) {

	projectID, ok := ProjectID(ctx)
	if !ok {
		return nil, fmt.Errorf("projectID missing in context")
	}

	// Resolve module
	mod, err := e.modules.GetModule(ctx, projectID, node.Uses, "")
	if err != nil {
		return nil, fmt.Errorf("module resolve failed: %w", err)
	}

	// Load HTTP spec
	spec, err := e.modules.GetStore().GetHttpSpec(ctx, mod.ID)
	if err != nil || spec == nil {
		return nil, fmt.Errorf("http spec not found for module %s", mod.Name)
	}

	// Render body template
	templateCtx := map[string]interface{}{
		"inputs": inputs,
		"steps":  ctx.Value("steps"),
	}

	var bodyTemplate map[string]interface{}
	if spec.BodyTemplate != nil {
		bodyTemplate = spec.BodyTemplate.AsMap()
	}

	bodyMap, err := RenderTemplate(bodyTemplate, templateCtx)
	if err != nil {
		return nil, fmt.Errorf("render template failed: %w", err)
	}

	bodyBytes, _ := json.Marshal(bodyMap)
	req, err := http.NewRequestWithContext(
		ctx,
		spec.Method,
		spec.Url,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}

	// Apply authentication
	if err := applyAuth(req, spec.Auth, inputs); err != nil {
		return nil, err
	}

	timeout := time.Duration(spec.TimeoutMs) * time.Millisecond
	e.client.Timeout = timeout

	var resp *http.Response
	err = Retry(int(spec.RetryCount), 200*time.Millisecond, func() error {
		resp, err = e.client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			return fmt.Errorf("retryable status %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var output map[string]interface{}
	if err := json.Unmarshal(respBytes, &output); err != nil {
		return nil, err
	}

	return output, nil
}
