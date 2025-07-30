package integration_tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	
	"sync"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	local_slack "github.com/dyluth/npc2/channels/slack"
	"github.com/dyluth/npc2/channels/api"
	"github.com/dyluth/npc2/middleware"
	"github.com/dyluth/npc2/npc"
)

// MockSocketModeClient is a mock for SocketModeClient interface.
type MockSocketModeClient struct {
	EventsChan chan socketmode.Event
}

func (m *MockSocketModeClient) Run() error {
	return nil
}

func (m *MockSocketModeClient) Ack(req socketmode.Request, payload ...interface{}) {
	// Do nothing for mock
}

func (m *MockSocketModeClient) Events() <-chan socketmode.Event {
	return m.EventsChan
}

func TestIntegration(t *testing.T) {
	// Set up environment variables for the main application
	os.Setenv("SLACK_APP_TOKEN", "xapp-test")
	os.Setenv("SLACK_BOT_TOKEN", "xoxb-test")
	os.Setenv("API_TOKEN", "test-api-token")

	// Create a new NPC core
	npcCore := npc.NewNpc()

	// Add authentication middleware
	authMiddleware := &middleware.AuthMiddleware{Token: os.Getenv("API_TOKEN")}
	npcCore.Use(authMiddleware)

	// Create and register a simple action
	helloAction := npc.Action{
		Name:        "hello",
		Description: "A simple hello action",
		Handler: func(request npc.Request) npc.Response {
			return npc.Response{Data: "Hello, world!", Code: 200}
		},
	}
	npcCore.RegisterAction(helloAction)

	// --- Test API Channel --- //
	apiChannel := api.NewAPIChannel(":8080")
	apiChannel.RegisterRequestHandler(npcCore.ProcessRequest)
	apiChannel.Start()
	defer apiChannel.Stop()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test API with valid token
	apiURL := "http://localhost:8080/api/request"
	requestBody, _ := json.Marshal(map[string]interface{}{"action": "hello"})
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", "Bearer test-api-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send API request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read API response body: %v", err)
	}
	t.Logf("API valid token response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("API request with valid token failed: status %d", resp.StatusCode)
	}

	if strings.TrimSpace(string(respBody)) != `"Hello, world!"` {
		t.Errorf("API response with valid token unexpected: %v", string(respBody))
	}

	// Test API with invalid token
	req, _ = http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", "Bearer wrong-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send API request with invalid token: %v", err)
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read API response body for invalid token: %v", err)
	}
	t.Logf("API invalid token response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("API request with invalid token expected unauthorized, got status %d", resp.StatusCode)
	}

	var errorResponse map[string]string
	json.Unmarshal(respBody, &errorResponse)
	if errorResponse["error"] != "unauthorized" {
		t.Errorf("API unauthorized response body unexpected: %q", errorResponse["error"])
	}

	// --- Test Slack Channel --- //
	// Create a new NPC core for Slack without auth middleware
	slackNpcCore := npc.NewNpc()
	slackNpcCore.RegisterAction(npc.Action{
		Name:        "hello",
		Description: "A simple hello action",
		Handler: func(request npc.Request) npc.Response {
			return npc.Response{Data: "Hello, world!", Code: 200}
		},
	})

	var wg sync.WaitGroup
	// Create a mock Slack client and socketmode client
	client := &slack.Client{}
	mockSocketMode := &MockSocketModeClient{
		EventsChan: make(chan socketmode.Event, 1),
	}

	// Create a new Slack channel with the mock clients
	slackChannel := &local_slack.SlackChannel{
		Client:     client,
		SocketMode: mockSocketMode,
	}

	// Register a mock request handler
	var handledSlackResponse npc.Response
	slackChannel.RegisterRequestHandler(func(request npc.Request) npc.Response {
		handledSlackResponse = slackNpcCore.ProcessRequest(request)
		wg.Done() // Signal that the request has been handled
		return handledSlackResponse
	})

	// Start the Slack channel
	slackChannel.Start()
	defer slackChannel.Stop()

	// Simulate an EventsAPI event
	wg.Add(1) // Increment wait group counter
	event := socketmode.Event{
		Type: socketmode.EventTypeEventsAPI,
		Data: slackevents.EventsAPIEvent{
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: slackevents.MessageEvent{
					Type: "message",
					Text: "hello",
					User: "U12345",
				},
			},
		},
		Request: &socketmode.Request{},
	}
	mockSocketMode.EventsChan <- event

	// Wait for the request to be handled
	wg.Wait()

	// Check if the request handler was called with the correct data
	if handledSlackResponse.Error != nil {
		t.Fatalf("Slack request handler returned an error: %v", handledSlackResponse.Error)
	}

	if handledSlackResponse.Data != "Hello, world!" {
		t.Errorf("Slack response data unexpected: %v", handledSlackResponse.Data)
	}
}
