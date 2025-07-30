package middleware

import (
	"log"

	"github.com/dyluth/npc2/npc"
)

// AuditLogMiddleware logs the action, source, and text of each request.
type AuditLogMiddleware struct{}

// Execute logs the request details.
func (m *AuditLogMiddleware) Execute(request *npc.Request) error {
	log.Printf("AUDIT: Action=%s, Source=%s, Text=%s, User=%s, ChannelID=%s",
		request.Action, request.Source, request.Text, request.User, request.ChannelID)
	return nil // Always continue the chain
}
