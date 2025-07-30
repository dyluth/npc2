package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyluth/npc2/channels/api"
	"github.com/dyluth/npc2/channels/slack"
	"github.com/dyluth/npc2/middleware"
	"github.com/dyluth/npc2/npc"
)

func main() {
	// Create a new NPC core
	npcCore := npc.NewNpc()

	// Get API token from environment variable
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		fmt.Println("API_TOKEN must be set.")
		return
	}

	// Create and register authentication middleware
	authMiddleware := &middleware.AuthMiddleware{Token: apiToken}
	npcCore.Use(authMiddleware)

	// Register audit logging middleware
	auditLogMiddleware := &middleware.AuditLogMiddleware{}
	npcCore.Use(auditLogMiddleware)

	// Create and register a simple action
	helloAction := npc.Action{
		Name:        "hello",
		Description: "A simple hello action",
		Handler: func(request npc.Request) npc.Response {
			return npc.Response{Data: "Hello, world!", Code: 200}
		},
	}
	npcCore.RegisterAction(helloAction)

	// Get Slack tokens from environment variables
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")

	if slackAppToken == "" || slackBotToken == "" {
		fmt.Println("SLACK_APP_TOKEN and SLACK_BOT_TOKEN must be set.")
		return
	}

	// Create and start the Slack channel
	slackChannel, err := slack.NewSlackChannel(slackAppToken, slackBotToken)
	if err != nil {
		fmt.Printf("Failed to create Slack channel: %v\n", err)
		return
	}
	slackChannel.RegisterRequestHandler(npcCore.ProcessRequest)
	slackChannel.Start()

	// Create and start the API channel
	apiChannel := api.NewAPIChannel(":8080")
	apiChannel.RegisterRequestHandler(npcCore.ProcessRequest)
	apiChannel.Start()

	fmt.Println("NPC is running. Press Ctrl+C to exit.")

	// Wait for a signal to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Stop the channels
	slackChannel.Stop()
	apiChannel.Stop()
}
