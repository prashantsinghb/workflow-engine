package executor

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/valyala/fasttemplate"
)

// flattenMap flattens a nested map into dot-notation keys for template access
// e.g., {"inputs": {"message": "test"}} -> {"inputs.message": "test"}
func flattenMap(prefix string, m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			// Recursively flatten nested maps
			for nestedKey, nestedVal := range flattenMap(key, val) {
				result[nestedKey] = nestedVal
			}
		case []interface{}:
			// Handle arrays by indexing
			for i, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					for nestedKey, nestedVal := range flattenMap(key+"."+strconv.Itoa(i), itemMap) {
						result[nestedKey] = nestedVal
					}
				} else {
					result[key+"."+strconv.Itoa(i)] = item
				}
			}
		default:
			result[key] = v
		}
	}
	return result
}

// RenderTemplate renders a JSON template using {{ }} placeholders.
// Context typically contains:
//
//	{
//	  "inputs": map[string]interface{},
//	  "steps":  map[string]interface{},
//	}
//
// Templates can access values using:
//   - {{message}} - direct access to top-level inputs
//   - {{inputs.message}} - nested access via dot notation
func RenderTemplate(
	tpl map[string]interface{},
	context map[string]interface{},
) (map[string]interface{}, error) {

	if tpl == nil {
		return nil, nil
	}

	// Flatten the context to support nested access like {{inputs.message}}
	flatContext := make(map[string]interface{})
	
	// First, add all top-level context values directly
	for k, v := range context {
		flatContext[k] = v
	}
	
	// Then, flatten nested structures (like inputs and steps) with dot notation
	if inputs, ok := context["inputs"].(map[string]interface{}); ok {
		for k, v := range flattenMap("inputs", inputs) {
			flatContext[k] = v
		}
		// Also add inputs directly for backward compatibility
		for k, v := range inputs {
			flatContext[k] = v
		}
	}
	
	if steps, ok := context["steps"].(map[string]interface{}); ok {
		for k, v := range flattenMap("steps", steps) {
			flatContext[k] = v
		}
	}

	raw, err := json.Marshal(tpl)
	if err != nil {
		return nil, fmt.Errorf("template marshal failed: %w", err)
	}

	// Convert flatContext values to strings for fasttemplate
	templateVars := make(map[string]interface{})
	for k, v := range flatContext {
		switch val := v.(type) {
		case string:
			templateVars[k] = val
		case int, int32, int64:
			templateVars[k] = fmt.Sprintf("%d", val)
		case float32, float64:
			templateVars[k] = fmt.Sprintf("%g", val)
		case bool:
			templateVars[k] = fmt.Sprintf("%t", val)
		case nil:
			templateVars[k] = ""
		default:
			// For complex types, marshal to JSON string
			if jsonBytes, err := json.Marshal(val); err == nil {
				templateVars[k] = string(jsonBytes)
			} else {
				templateVars[k] = fmt.Sprintf("%v", val)
			}
		}
	}

	t := fasttemplate.New(string(raw), "{{", "}}")
	rendered := t.ExecuteString(templateVars)

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(rendered), &out); err != nil {
		return nil, fmt.Errorf("template unmarshal failed: %w", err)
	}

	return out, nil
}
