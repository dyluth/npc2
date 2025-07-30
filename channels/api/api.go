package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dyluth/npc2/npc"
)

// APIChannel is a communication channel for a REST API.
type APIChannel struct {
	port           string
	server         *http.Server
	requestHandler func(request npc.Request) npc.Response
}

// NewAPIChannel creates a new APIChannel instance.
func NewAPIChannel(port string) *APIChannel {
	return &APIChannel{
		port: port,
	}
}

// Start starts the API communication channel.
func (ac *APIChannel) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/request", ac.handleRequest)

	ac.server = &http.Server{
		Addr:    ac.port,
		Handler: mux,
	}

	go func() {
		if err := ac.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server failed: %v", err)
		}
	}()
	fmt.Printf("API server listening on port %s\n", ac.port)
}

// Stop stops the API communication channel.
func (ac *APIChannel) Stop() {
	if err := ac.server.Shutdown(context.Background()); err != nil {
		log.Printf("API server shutdown failed: %v", err)
	}
	fmt.Println("API server stopped.")
}

// SendMessage is not applicable for the API channel in this context.
func (ac *APIChannel) SendMessage(channelID string, message string) {
	log.Println("SendMessage is not implemented for the API channel")
}

// RegisterRequestHandler registers a handler for incoming requests.
func (ac *APIChannel) RegisterRequestHandler(handler func(request npc.Request) npc.Response) {
	ac.requestHandler = handler
}

func (ac *APIChannel) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request method"})
		return
	}

	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON payload"})
		return
	}

	// Extract action, user, and channel from requestData
	actionName, _ := requestData["action"].(string)
	user := ""
	channelID := ""

	// Extract args
	args := make(map[string]string)
	if rawArgs, ok := requestData["args"].(map[string]interface{}); ok {
		for k, v := range rawArgs {
			if strVal, isString := v.(string); isString {
				args[k] = strVal
			}
		}
	}

	// If the API request includes user/channel_id, extract them
	if u, ok := requestData["user"].(string); ok {
		user = u
	}
	if c, ok := requestData["channel_id"].(string); ok {
		channelID = c
	}

	// Extract text representation of the payload
	textPayload := ""
	if msg, ok := requestData["message"].(string); ok {
		textPayload = msg
	} else if action, ok := requestData["action"].(string); ok {
		textPayload = action // Use action as text if no message field
	}

	// Extract authentication information
	authMethod := "apikey"
	authToken := r.Header.Get("Authorization")
	if strings.HasPrefix(authToken, "Bearer ") {
		authToken = strings.TrimPrefix(authToken, "Bearer ")
	}

	// Construct npc.Request
	npcRequest := npc.Request{
		Action:    actionName,
		User:      user,
		ChannelID: channelID,
		Text:      textPayload,
		Source:    "API",
		AuthMethod: authMethod,
		AuthToken: authToken,
		Args:      args,
		RawData:   requestData,
	}

	if ac.requestHandler != nil {
		response := ac.requestHandler(npcRequest)

		if response.Error != nil {
			w.Header().Set("Content-Type", "application/json")
			if response.Error.Error() == "unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			json.NewEncoder(w).Encode(map[string]string{"error": response.Error.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		log.Printf("API response type before encoding: %T, value: %v", response.Data, response.Data)
		if err := json.NewEncoder(w).Encode(response.Data); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "No request handler registered"})
	}
}
