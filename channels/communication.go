package channels

// Communication defines the interface for communication channels.
type Communication interface {
	Start()
	Stop()
	SendMessage(channelID string, message string)
	RegisterRequestHandler(handler func(request interface{}) interface{})
}
