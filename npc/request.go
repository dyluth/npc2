package npc

// Request encapsulates a standardized incoming request.
type Request struct {
	Action    string
	User      string
	ChannelID string
	Text      string      // Textual representation of the payload
	Source    string      // e.g., "API", "Slack"
	AuthMethod string     // e.g., "apikey", "slack_user"
	AuthToken string      // The actual token or user ID
	Args      map[string]string // Arbitrary key-value arguments
	RawData   interface{}
}
