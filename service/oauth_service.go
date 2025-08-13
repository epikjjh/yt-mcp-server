package service

// OAuthService is a simplified service for a single-user setup.
// It primarily provides the necessary OAuth discovery metadata.
// In this simplified model, we don't need dynamic client registration
// or complex token management for the MCP client itself.
type OAuthService struct {
	serverURL string
}

// NewOAuthService creates a new simplified OAuthService.
func NewOAuthService(serverURL string) *OAuthService {
	return &OAuthService{
		serverURL: serverURL,
	}
}

// GetAuthServerMetadata returns OAuth authorization server metadata (RFC 8414).
// For our simplified case, many of these endpoints won't be fully implemented
// as we bypass the MCP-level OAuth flow in favor of a direct Google login check.
func (s *OAuthService) GetAuthServerMetadata() map[string]interface{} {
	return map[string]interface{}{
		"issuer":                   s.serverURL,
		"authorization_endpoint":   s.serverURL + "/oauth/authorize",
		"token_endpoint":           s.serverURL + "/oauth/token", // This will not be used
		"scopes_supported":         []string{"youtube:transcripts"},
		"response_types_supported": []string{"code"},
		"grant_types_supported":    []string{"authorization_code"},
	}
}

// GetProtectedResourceMetadata returns resource server metadata (RFC 9728).
func (s *OAuthService) GetProtectedResourceMetadata() map[string]interface{} {
	return map[string]interface{}{
		"resource":              s.serverURL,
		"authorization_servers": []string{s.serverURL},
	}
}
