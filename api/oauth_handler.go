package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yt-mcp-server/service"
)

type OAuthHandler struct {
	oauthService  *service.OAuthService
	googleService *service.GoogleOAuthService
}

func NewOAuthHandler(oauthService *service.OAuthService, googleService *service.GoogleOAuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService:  oauthService,
		googleService: googleService,
	}
}

// Authorize handles the start of the OAuth flow.
// In our simplified single-user model, this just redirects to Google.
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	// The state parameter is not strictly needed for a single-user app, but it's good practice.
	googleAuthURL := h.googleService.GetAuthURL("state-token")
	h.renderAuthorizePage(w, googleAuthURL)
}

// GoogleCallback handles the redirect from Google after the user grants consent.
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		http.Error(w, fmt.Sprintf("OAuth error from Google: %s", errorParam), http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "Missing authorization code from Google", http.StatusBadRequest)
		return
	}

	if err := h.googleService.ExchangeCodeForToken(r.Context(), code); err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange Google code for token: %v", err), http.StatusInternalServerError)
		return
	}

	// Render a simple success page.
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Authentication Successful</title>
		<style>body{font-family: Arial, sans-serif; text-align: center; margin-top: 100px;}</style>
	</head>
	<body>
		<h1>✅ Authentication Successful!</h1>
		<p>You can now close this window and return to your application.</p>
	</body>
	</html>
	`)
}

// RequireAuth is a middleware that checks if the user is authenticated with Google.
func (h *OAuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.googleService.IsAuthenticated() {
			h.sendUnauthorized(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// sendUnauthorized sends a 401 response, prompting the user to authenticate.
func (h *OAuthHandler) sendUnauthorized(w http.ResponseWriter, r *http.Request) {
	// In this simplified model, we don't need complex WWW-Authenticate headers.
	// We can just return a simple error message.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             "unauthorized",
		"error_description": "Not authenticated with Google. Please visit /oauth/authorize to log in.",
	})
}

// renderAuthorizePage renders the page that prompts the user to log in.
func (h *OAuthHandler) renderAuthorizePage(w http.ResponseWriter, googleAuthURL string) {
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>YouTube MCP Server - Authorization</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; }
			.button { 
				display: inline-block; 
				padding: 12px 24px; 
				background-color: #4285f4; 
				color: white; 
				text-decoration: none; 
				border-radius: 4px; 
				margin: 20px 0;
			}
			.button:hover { background-color: #3367d6; }
		</style>
	</head>
	<body>
		<h1>YouTube MCP Server Authorization</h1>
		<p>This server needs permission to access YouTube transcripts on your behalf.</p>
		<p>Click the button below to sign in with your Google account.</p>
		<a href="{{.GoogleAuthURL}}" class="button">Authorize with Google</a>
	</body>
	</html>
	`

	t, err := template.New("authorize").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, map[string]interface{}{"GoogleAuthURL": googleAuthURL})
}

// SetupOAuthRoutes sets up the simplified OAuth-related routes.
func SetupOAuthRoutes(r chi.Router, handler *OAuthHandler) {
	// These discovery endpoints are not strictly necessary for our simplified flow,
	// but we keep them for potential future compatibility.
	r.Get("/.well-known/oauth-protected-resource", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(handler.oauthService.GetProtectedResourceMetadata())
	})
	r.Get("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(handler.oauthService.GetAuthServerMetadata())
	})

	// OAuth flow endpoints
	r.Get("/oauth/authorize", handler.Authorize)
	r.Get("/oauth/callback", handler.GoogleCallback)
}
