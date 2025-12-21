package executor

import (
	"fmt"
	"net/http"

	"github.com/prashantsinghb/workflow-engine/api/service"
)

func applyAuth(
	req *http.Request,
	auth *service.HttpAuth,
	inputs map[string]interface{},
) error {

	if auth == nil || auth.Type == nil {
		return nil
	}

	switch v := auth.Type.(type) {
	case *service.HttpAuth_Bearer:
		if v.Bearer != nil && v.Bearer.Token != "" {
			req.Header.Set("Authorization", "Bearer "+v.Bearer.Token)
		}
	case *service.HttpAuth_ApiKey:
		if v.ApiKey != nil && v.ApiKey.Header != "" && v.ApiKey.Value != "" {
			req.Header.Set(v.ApiKey.Header, v.ApiKey.Value)
		}
	case *service.HttpAuth_Oauth2:
		// OAuth2 requires token exchange - simplified for now
		if v.Oauth2 != nil && v.Oauth2.TokenUrl != "" {
			// TODO: Implement OAuth2 token exchange
			return fmt.Errorf("OAuth2 authentication not yet implemented")
		}
	default:
		return fmt.Errorf("unsupported auth type")
	}
	return nil
}
