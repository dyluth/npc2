package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dyluth/npc2/npc"
)

// TestAPIChannel tests the APIChannel.
func TestAPIChannel(t *testing.T) {
	// Create a new API channel
	apiChannel := NewAPIChannel(":8081")

	// Register a mock request handler
	apiChannel.RegisterRequestHandler(func(request npc.Request) npc.Response {
		// Verify the incoming request object
		if request.Action != "test" {
			t.Errorf("Expected action 'test', got %s", request.Action)
		}
		if request.Source != "API" {
			t.Errorf("Expected source 'API', got %s", request.Source)
		}
		if request.AuthMethod != "apikey" {
			t.Errorf("Expected auth method 'apikey', got %s", request.AuthMethod)
		}
		if request.AuthToken != "test-token" {
			t.Errorf("Expected auth token 'test-token', got %s", request.AuthToken)
		}
		if request.Text != "test message" {
			t.Errorf("Expected text 'test message', got %s", request.Text)
		}
		if request.Args["key1"] != "value1" {
			t.Errorf("Expected arg key1 'value1', got %s", request.Args["key1"])
		}
		if request.Args["key2"] != "value2" {
			t.Errorf("Expected arg key2 'value2', got %s", request.Args["key2"])
		}

		return npc.Response{Data: "ok", Code: 200}
	})

	// Start the API channel
	apiChannel.Start()
	defer apiChannel.Stop()

	// Create a test request
	requestBody, _ := json.Marshal(map[string]interface{}{"action": "test", "message": "test message", "args": map[string]string{"key1": "value1", "key2": "value2"}})
	req, err := http.NewRequest("POST", "/api/request", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")

	// Create a test recorder
	rr := httptest.NewRecorder()

	// Handle the request
	apiChannel.handleRequest(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := `"ok"`
	actual := strings.TrimSpace(rr.Body.String())
	if actual != expected {
		t.Errorf("handler returned unexpected body: got %q want %q",
			actual, expected)
	}
}
