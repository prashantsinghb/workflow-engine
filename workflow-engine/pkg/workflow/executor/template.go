package executor

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasttemplate"
)

// RenderTemplate renders a JSON template using {{ }} placeholders.
// Context typically contains:
//
//	{
//	  "inputs": map[string]interface{},
//	  "steps":  map[string]interface{},
//	}
func RenderTemplate(
	tpl map[string]interface{},
	context map[string]interface{},
) (map[string]interface{}, error) {

	if tpl == nil {
		return nil, nil
	}

	raw, err := json.Marshal(tpl)
	if err != nil {
		return nil, fmt.Errorf("template marshal failed: %w", err)
	}

	t := fasttemplate.New(string(raw), "{{", "}}")
	rendered := t.ExecuteString(context)

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(rendered), &out); err != nil {
		return nil, fmt.Errorf("template unmarshal failed: %w", err)
	}

	return out, nil
}
