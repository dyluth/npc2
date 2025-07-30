package npc

import "fmt"

// Middleware defines the interface for middleware components.
type Middleware interface {
	Execute(request *Request) error
}

// Action defines the structure for bot actions.
type Action struct {
	Name        string
	Description string
	Handler     func(Request) Response
}

// Npc is the core bot engine.
type Npc struct {
	actions    map[string]Action
	middleware []Middleware
}

// NewNpc creates a new Npc instance.
func NewNpc() *Npc {
	return &Npc{
		actions:    make(map[string]Action),
		middleware: make([]Middleware, 0),
	}
}

// RegisterAction adds a new action to the bot.
func (n *Npc) RegisterAction(action Action) {
	n.actions[action.Name] = action
}

// Use adds a new middleware to the pipeline.
func (n *Npc) Use(middleware Middleware) {
	n.middleware = append(n.middleware, middleware)
}

// ProcessRequest processes a request by executing the middleware chain and then the appropriate action.
func (n *Npc) ProcessRequest(request Request) Response {
	currentRequest := &request // Pass a pointer to the request

	for _, m := range n.middleware {
		err := m.Execute(currentRequest)
		if err != nil {
			return Response{Error: err} // Stop chain on error
		}
	}

	// After all middleware, execute the action
	// If the action is not found, return an error response.
	if action, ok := n.actions[currentRequest.Action]; ok {
		return action.Handler(*currentRequest) // Pass the dereferenced request to the handler
	}
	return Response{Error: fmt.Errorf("action %s not found", currentRequest.Action)}
}
