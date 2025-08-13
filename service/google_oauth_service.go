package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// GoogleOAuthService handles the Google OAuth2 flow and token management.
type GoogleOAuthService struct {
	tokenStore  *InMemoryTokenStore
	oauthConfig *oauth2.Config
}

// NewGoogleOAuthService creates a new GoogleOAuthService.
func NewGoogleOAuthService(tokenStore *InMemoryTokenStore, clientID, clientSecret, redirectURI string) *GoogleOAuthService {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/youtube.force-ssl",
			"https://www.googleapis.com/auth/youtubepartner",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuthService{
		tokenStore:  tokenStore,
		oauthConfig: config,
	}
}

// GetAuthURL returns the Google OAuth authorization URL.
func (s *GoogleOAuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCodeForToken exchanges an authorization code for a token and stores it.
func (s *GoogleOAuthService) ExchangeCodeForToken(ctx context.Context, code string) error {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}
	s.tokenStore.SetToken(token)
	log.Println("✅ Successfully authenticated with Google and stored token in memory.")
	return nil
}

// GetYouTubeService returns an authenticated YouTube service client.
// It uses the stored token and handles automatic refresh.
func (s *GoogleOAuthService) GetYouTubeService(ctx context.Context) (*youtube.Service, error) {
	token := s.tokenStore.GetToken()
	if token == nil {
		return nil, fmt.Errorf("not authenticated with Google; please visit /oauth/authorize")
	}

	tokenSource := s.oauthConfig.TokenSource(ctx, token)
	
	// The TokenSource will automatically refresh the token if it's expired.
	// We can check if it was refreshed and update our store.
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve or refresh token: %w", err)
	}

	// If the access token is new, update the store.
	if newToken.AccessToken != token.AccessToken {
		s.tokenStore.SetToken(newToken)
		log.Println("♻️ Google OAuth token was refreshed.")
	}

	youtubeService, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	return youtubeService, nil
}

// IsAuthenticated checks if a valid token is currently stored.
func (s *GoogleOAuthService) IsAuthenticated() bool {
	return s.tokenStore.GetToken() != nil
}

// TokenRefresher is a background goroutine that proactively refreshes the token.
func (s *GoogleOAuthService) TokenRefresher(ctx context.Context) {
	log.Println("background token refresher started")
	ticker := time.NewTicker(15 * time.Minute) // Check every 15 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			token := s.tokenStore.GetToken()
			if token == nil {
				continue // No token to refresh
			}

			// Proactively refresh if the token will expire in the next 30 minutes.
			if !token.Valid() || token.Expiry.Before(time.Now().Add(30*time.Minute)) {
				log.Println("Proactively refreshing token...")
				tokenSource := s.oauthConfig.TokenSource(ctx, token)
				newToken, err := tokenSource.Token()
				if err != nil {
					log.Printf("Error refreshing token in background: %v", err)
					// If refresh fails, the token might be invalid. Clear it.
					s.tokenStore.SetToken(nil)
					continue
				}
				s.tokenStore.SetToken(newToken)
				log.Println("✅ Token proactively refreshed in the background.")
			}
		}
	}
}
