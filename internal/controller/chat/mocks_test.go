package chatctrl

import (
	"errors"
	"sync"
)

var ErrConnectionClosed = errors.New("connection closed")

// MockWebSocketConn is a mock implementation for testing broadcast and client management patterns
type MockWebSocketConn struct {
	mu         sync.Mutex
	messages   []string
	binaryMsgs [][]byte
	closed     bool
	ID         string
}

// NewMockWebSocketConn creates a new mock connection with a given ID
func NewMockWebSocketConn(id string) *MockWebSocketConn {
	return &MockWebSocketConn{
		ID:       id,
		messages: []string{},
	}
}

// SendText sends a text message (mock implementation)
func (m *MockWebSocketConn) SendText(msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrConnectionClosed
	}
	m.messages = append(m.messages, msg)
	return nil
}

// SendBinary sends a binary message (mock implementation)
func (m *MockWebSocketConn) SendBinary(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrConnectionClosed
	}
	m.binaryMsgs = append(m.binaryMsgs, data)
	return nil
}

// GetMessages returns all text messages sent to this connection
func (m *MockWebSocketConn) GetMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to prevent external modifications
	result := make([]string, len(m.messages))
	copy(result, m.messages)
	return result
}

// GetBinaryMessages returns all binary messages sent to this connection
func (m *MockWebSocketConn) GetBinaryMessages() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([][]byte, len(m.binaryMsgs))
	copy(result, m.binaryMsgs)
	return result
}

// Close marks the connection as closed
func (m *MockWebSocketConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrConnectionClosed
	}
	m.closed = true
	return nil
}

// IsClosed returns whether the connection is closed
func (m *MockWebSocketConn) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// Clear clears all messages (useful for test setup)
func (m *MockWebSocketConn) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = []string{}
	m.binaryMsgs = [][]byte{}
}

// MockClientManager is used for testing client broadcast and management
type MockClientManager struct {
	mu      sync.RWMutex
	clients map[string]*MockWebSocketConn
}

// NewMockClientManager creates a new mock client manager
func NewMockClientManager() *MockClientManager {
	return &MockClientManager{
		clients: make(map[string]*MockWebSocketConn),
	}
}

// Add adds a mock client
func (m *MockClientManager) Add(id string, conn *MockWebSocketConn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[id] = conn
}

// Remove removes a mock client
func (m *MockClientManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, id)
}

// Broadcast sends a message to all connected clients
func (m *MockClientManager) Broadcast(msg string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, conn := range m.clients {
		if err := conn.SendText(msg); err == nil {
			count++
		}
	}
	return count
}

// BroadcastExcept sends a message to all clients except the specified one
func (m *MockClientManager) BroadcastExcept(exceptID, msg string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for id, conn := range m.clients {
		if id != exceptID {
			if err := conn.SendText(msg); err == nil {
				count++
			}
		}
	}
	return count
}

// GetClientCount returns the number of connected clients
func (m *MockClientManager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// GetClient retrieves a specific client by ID
func (m *MockClientManager) GetClient(id string) *MockWebSocketConn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[id]
}
