package slack

import (
	"sync"
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

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

// TestSlackChannel tests the SlackChannel.
func TestSlackChannel(t *testing.T) {
	var wg sync.WaitGroup

	// Create a mock Slack client and socketmode client
	client := &slack.Client{}
	mockSocketMode := &MockSocketModeClient{
		EventsChan: make(chan socketmode.Event, 1),
	}

	// Create a new Slack channel with the mock clients
	sc := &SlackChannel{
		Client:     client,
		SocketMode: mockSocketMode,
	}

	// Register a mock request handler
	var handledRequest npc.Request
	sc.RegisterRequestHandler(func(request npc.Request) npc.Response {
		handledRequest = request
		wg.Done() // Signal that the request has been handled
		return npc.Response{Data: "ok", Code: 200}
	})

	// Start the Slack channel
	sc.Start()

	// Simulate an EventsAPI event
	wg.Add(1) // Increment wait group counter
	event := socketmode.Event{
		Type: socketmode.EventTypeEventsAPI,
		Data: slackevents.EventsAPIEvent{
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: slackevents.MessageEvent{
					Type:    "message",
					Text:    "hello",
					User:    "U12345",
					Channel: "C12345",
					ChannelType: "channel",
				},
			},
		},
		Request: &socketmode.Request{},
	}
	mockSocketMode.EventsChan <- event

	// Wait for the request to be handled
	wg.Wait()

	// Check if the request handler was called with the correct data
	if handledRequest.Action != "hello" {
		t.Errorf("Expected action 'hello', got %s", handledRequest.Action)
	}
	if handledRequest.User != "U12345" {
		t.Errorf("Expected user 'U12345', got %s", handledRequest.User)
	}
	if handledRequest.ChannelID != "C12345" {
		t.Errorf("Expected channel ID 'C12345', got %s", handledRequest.ChannelID)
	}
	if handledRequest.Text != "hello" {
		t.Errorf("Expected text 'hello', got %s", handledRequest.Text)
	}
	if handledRequest.Source != "Slack" {
		t.Errorf("Expected source 'Slack', got %s", handledRequest.Source)
	}
	if handledRequest.AuthMethod != "slack_user" {
		t.Errorf("Expected auth method 'slack_user', got %s", handledRequest.AuthMethod)
	}
	if handledRequest.AuthToken != "U12345" {
		t.Errorf("Expected auth token 'U12345', got %s", handledRequest.AuthToken)
	}

	if handledRequest.Args["channel_type"] != "channel" {
		t.Errorf("Expected Args[\"channel_type\"] to be 'channel', got %s", handledRequest.Args["channel_type"])
	}

	if _, ok := handledRequest.RawData.(slackevents.MessageEvent); !ok {
		t.Errorf("Expected RawData to be of type slackevents.MessageEvent, got %T", handledRequest.RawData)
	}
}
