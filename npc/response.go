package npc

// Response encapsulates the result of an action or middleware execution.
type Response struct {
	Data  string
	Error error
	Code  int
}
