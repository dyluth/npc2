package slack

import (
	"fmt"
	"log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/dyluth/npc2/npc"
)

// SocketModeClient is an interface for the socketmode.Client to allow mocking.
type SocketModeClient interface {
	Run() error
	Ack(req socketmode.Request, payload ...interface{})
	Events() <-chan socketmode.Event
}

// realSocketModeClient implements the SocketModeClient interface for the actual slack socketmode client.
type realSocketModeClient struct {
	*socketmode.Client
}

func (rsmc *realSocketModeClient) Events() <-chan socketmode.Event {
	return rsmc.Client.Events
}

// SlackChannel is a communication channel for Slack.
type SlackChannel struct {
	Client               *slack.Client
	SocketMode           SocketModeClient
	requestHandler       func(request npc.Request) npc.Response
}

// NewSlackChannel creates a new SlackChannel instance.
func NewSlackChannel(appToken, botToken string) (*SlackChannel, error) {
	client := slack.New(botToken, slack.OptionAppLevelToken(appToken))
	socketModeClient := socketmode.New(client)

	return &SlackChannel{
		Client:     client,
		SocketMode: &realSocketModeClient{socketModeClient},
	}, nil
}

// Start starts the Slack communication channel.
func (sc *SlackChannel) Start() {
	go func() {
		for evt := range sc.SocketMode.Events() {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("Connecting to Slack...")
			case socketmode.EventTypeConnectionError:
				fmt.Println("Connection failed. Retrying...")
			case socketmode.EventTypeConnected:
				fmt.Println("Connected to Slack.")
			case socketmode.EventTypeEventsAPI:
				sc.handleEvent(evt)
			}
		}
	}()

	go func() {
		if err := sc.SocketMode.Run(); err != nil {
			log.Fatalf("Socket mode run failed: %v", err)
		}
	}()
}

// Stop stops the Slack communication channel.
func (sc *SlackChannel) Stop() {
	// The slack-go library doesn't provide a direct way to stop the socket mode client.
	// In a real application, you might need to manage the lifecycle more carefully.
	fmt.Println("Slack channel stopping...")
}

// SendMessage sends a message to a Slack channel.
func (sc *SlackChannel) SendMessage(channelID string, message string) {
	_, _, err := sc.Client.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("Failed to send message to channel %s: %v", channelID, err)
	}
}

// RegisterRequestHandler registers a handler for incoming requests.
func (sc *SlackChannel) RegisterRequestHandler(handler func(request npc.Request) npc.Response) {
	sc.requestHandler = handler
}

func (sc *SlackChannel) handleEvent(evt socketmode.Event) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return
	}

	sc.SocketMode.Ack(*evt.Request)

	if sc.requestHandler != nil {
		if messageEvent, ok := eventsAPIEvent.InnerEvent.Data.(slackevents.MessageEvent); ok {
			action := "unknown" // Default action
			args := make(map[string]string)
			if messageEvent.Text == "hello" {
				action = "hello"
				// Example of adding args from Slack message
				args["channel_type"] = messageEvent.ChannelType
			}

			// Construct npc.Request
			npcRequest := npc.Request{
				Action:    action,
				User:      messageEvent.User,
				ChannelID: messageEvent.Channel,
				Text:      messageEvent.Text, // Populate Text field
				Source:    "Slack",           // Set Source
				AuthMethod: "slack_user",    // Set AuthMethod
				AuthToken: messageEvent.User, // Set AuthToken
				Args:      args,
				RawData:   messageEvent,      // Store the original message event
			}

			sc.requestHandler(npcRequest)
		} else {
			// For other event types, create a generic request
			npcRequest := npc.Request{
				Action:    "unknown", // Default action for non-message events
				Source:    "Slack",
				AuthMethod: "none", // Or appropriate default
				AuthToken: "",
				Args:      make(map[string]string), // Initialize empty Args map
				RawData:   eventsAPIEvent.InnerEvent.Data,
			}
			sc.requestHandler(npcRequest)
		}
	}
}
