package service

import (
	"sync"

	"golang.org/x/oauth2"
)

// InMemoryTokenStore holds the token in memory for a single user.
// A mutex is used to handle concurrent access safely.
type InMemoryTokenStore struct {
	mu    sync.RWMutex
	token *oauth2.Token
}

// GetToken retrieves the token from the store.
func (ts *InMemoryTokenStore) GetToken() *oauth2.Token {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.token == nil {
		return nil
	}
	// Return a copy to avoid race conditions if the caller modifies it
	tokenCopy := *ts.token
	return &tokenCopy
}

// SetToken saves the token to the store.
func (ts *InMemoryTokenStore) SetToken(token *oauth2.Token) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.token = token
}
