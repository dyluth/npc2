package middleware

import (
	"net/http"
	"testing"

	"github.com/dyluth/npc2/npc"
)

// TestAuthMiddleware tests the AuthMiddleware.
func TestAuthMiddleware(t *testing.T) {
	middleware := &AuthMiddleware{Token: "test-token"}

	// Test with a valid token
	request := npc.Request{
		AuthMethod: "apikey",
		AuthToken:  "test-token",
		RawData:    map[string]interface{}{"headers": http.Header{"Authorization": []string{"Bearer test-token"}}},
	}

	err := middleware.Execute(&request)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with an invalid token
	request = npc.Request{
		AuthMethod: "apikey",
		AuthToken:  "invalid-token",
		RawData:    map[string]interface{}{"headers": http.Header{"Authorization": []string{"Bearer invalid-token"}}},
	}
	err = middleware.Execute(&request)

	if err == nil || err.Error() != "unauthorized" {
		t.Errorf("Expected 'unauthorized' error, got %v", err)
	}

	// Test with no token
	request = npc.Request{
		AuthMethod: "apikey",
		AuthToken:  "",
		RawData:    map[string]interface{}{"headers": http.Header{}},
	}
	err = middleware.Execute(&request)

	if err == nil || err.Error() != "unauthorized" {
		t.Errorf("Expected 'unauthorized' error, got %v", err)
	}
}
