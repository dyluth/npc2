package npc

import (
	"testing"
)

// MockMiddleware is a mock middleware for testing.
type MockMiddleware struct {
	executed bool
}

// Execute marks the middleware as executed.
func (m *MockMiddleware) Execute(request *Request) error {
	m.executed = true
	return nil
}

// TestNewNpc tests the NewNpc function.
func TestNewNpc(t *testing.T) {
	npc := NewNpc()
	if npc == nil {
		t.Error("NewNpc() returned nil")
	}
	if npc.actions == nil {
		t.Error("NewNpc() did not initialize actions map")
	}
	if npc.middleware == nil {
		t.Error("NewNpc() did not initialize middleware slice")
	}
}

// TestRegisterAction tests the RegisterAction method.
func TestRegisterAction(t *testing.T) {
	npc := NewNpc()
	action := Action{
		Name: "test",
		Handler: func(request Request) Response {
			return Response{Data: "test"}
		},
	}
	npc.RegisterAction(action)
	if _, ok := npc.actions["test"]; !ok {
		t.Error("RegisterAction() did not register the action")
	}
}

// TestUse tests the Use method.
func TestUse(t *testing.T) {
	npc := NewNpc()
	middleware := &MockMiddleware{}
	npc.Use(middleware)
	if len(npc.middleware) != 1 {
		t.Error("Use() did not add the middleware")
	}
}

// TestProcessRequest tests the ProcessRequest method.
func TestProcessRequest(t *testing.T) {
	npc := NewNpc()
	action := Action{
		Name: "test",
		Handler: func(request Request) Response {
			return Response{Data: "test"}
		},
	}
	npc.RegisterAction(action)

	middleware := &MockMiddleware{}
	npc.Use(middleware)

	request := Request{
		Action: "test",
	}

	response := npc.ProcessRequest(request)

	if !middleware.executed {
		t.Error("ProcessRequest() did not execute the middleware")
	}

	if response.Error != nil {
		t.Errorf("ProcessRequest() returned an error: %v", response.Error)
	}

	if response.Data != "test" {
		t.Errorf("ProcessRequest() returned %v, expected %v", response.Data, "test")
	}

	// Test with an action not found
	request = Request{
		Action: "nonexistent",
	}
	response = npc.ProcessRequest(request)
	if response.Error == nil || response.Error.Error() != "action nonexistent not found" {
		t.Errorf("Expected 'action nonexistent not found' error, got %v", response.Error)
	}
}

