package middleware

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/dyluth/npc2/npc"
)

func TestAuditLogMiddleware(t *testing.T) {
	middleware := &AuditLogMiddleware{}

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0) // Disable all output flags, including timestamp
	defer func() { 
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags) // Restore default flags
	}() // Restore default output

	request := &npc.Request{
		Action:    "test_action",
		Source:    "test_source",
		Text:      "test_text",
		User:      "test_user",
		ChannelID: "test_channel",
	}

	err := middleware.Execute(request)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLog := "AUDIT: Action=test_action, Source=test_source, Text=test_text, User=test_user, ChannelID=test_channel\n"
	if buf.String() != expectedLog {
		t.Errorf("Expected log %q, got %q", expectedLog, buf.String())
	}
}
