package middleware

import (
	"fmt"

	"github.com/dyluth/npc2/npc"
)

// AuthMiddleware is a middleware for authenticating requests.
type AuthMiddleware struct {
	Token string
}

// Execute executes the authentication middleware.
func (m *AuthMiddleware) Execute(request *npc.Request) error {
	if request.AuthMethod == "apikey" && request.AuthToken == m.Token {
		return nil // Authentication successful, continue to next middleware
	}
	return fmt.Errorf("unauthorized")
}
