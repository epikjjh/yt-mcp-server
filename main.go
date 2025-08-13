package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/yt-mcp-server/api"
	"github.com/yt-mcp-server/config"
	"github.com/yt-mcp-server/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create an in-memory token store for our single user.
	tokenStore := &service.InMemoryTokenStore{}

	// Initialize services
	oauthService := service.NewOAuthService(cfg.MCPServerURL)
	googleService := service.NewGoogleOAuthService(tokenStore, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURI)
	youtubeService := service.NewYouTubeService(googleService)

	// Start the proactive token refresher
	go googleService.TokenRefresher(context.Background())

	// Initialize handlers
	oauthHandler := api.NewOAuthHandler(oauthService, googleService)
	mcpHandler := api.NewMCPHandler(youtubeService)

	// Setup HTTP router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://claude.ai", "https://*.claude.ai", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// OAuth routes
	api.SetupOAuthRoutes(r, oauthHandler)

	// MCP endpoint (protected)
	r.Route("/mcp", func(r chi.Router) {
		r.Use(oauthHandler.RequireAuth)
		r.Post("/", mcpHandler.HandleMCP)
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","server":"youtube-mcp-server","version":"1.0.0"}`)
	})

	// Root endpoint with server info
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>YouTube MCP Server</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 40px; }
        .section { margin: 30px 0; }
        .code { background-color: #f5f5f5; padding: 10px; border-radius: 4px; font-family: monospace; }
        .endpoint { margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YouTube MCP Server</h1>
        <p>OAuth-enabled Model Context Protocol server for YouTube transcript access</p>
    </div>

    <div class="section">
        <h2>Server Status</h2>
        <p>✅ Server is running and healthy</p>
        <p>🔐 OAuth 2.1 authentication enabled</p>
        <p>📺 YouTube Data API v3 integration active</p>
    </div>

    <div class="section">
        <h2>Available Endpoints</h2>
        <div class="endpoint">
            <strong>OAuth Discovery:</strong>
            <div class="code">GET /.well-known/oauth-protected-resource</div>
            <div class="code">GET /.well-known/oauth-authorization-server</div>
        </div>
        <div class="endpoint">
            <strong>OAuth Flow:</strong>
            <div class="code">POST /oauth/register</div>
            <div class="code">GET /oauth/authorize</div>
            <div class="code">POST /oauth/token</div>
        </div>
        <div class="endpoint">
            <strong>MCP Protocol:</strong>
            <div class="code">POST /mcp (requires authentication)</div>
        </div>
    </div>

    <div class="section">
        <h2>Claude.ai Integration</h2>
        <p>To use this server with Claude.ai, add the following to your MCP configuration:</p>
        <div class="code">
{<br>
  &nbsp;&nbsp;"youtube-transcript": {<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"command": "mcp",<br>
  &nbsp;&nbsp;&nbsp;&nbsp;"args": ["--server", "%s"]<br>
  &nbsp;&nbsp;}<br>
}
        </div>
    </div>

    <div class="section">
        <h2>Available Tools</h2>
        <ul>
            <li><strong>get_transcript</strong> - Get YouTube video transcripts with OAuth authentication</li>
        </ul>
    </div>
</body>
</html>
`, cfg.MCPServerURL)
	})

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Received shutdown signal, gracefully shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("🚀 YouTube MCP Server starting on port %s", cfg.Port)
	log.Printf("🔗 Server URL: %s", cfg.MCPServerURL)
	log.Printf("📋 OAuth Discovery: %s/.well-known/oauth-protected-resource", cfg.MCPServerURL)
	log.Printf("🎥 Ready to serve YouTube transcripts with OAuth authentication!")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}


