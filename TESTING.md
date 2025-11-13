# Testing WebSocket Controllers - Quick Reference

## Overview
This guide demonstrates how to test WebSocket controllers in the i-chat-server project using unit tests, integration tests, and mock objects.

## Test Files Created

### 1. `controller_test.go` - Unit Tests
**Location**: `internal/controller/chat/controller_test.go`

Tests controller initialization and the option pattern:
- `TestNewChatControllerWithoutOptions`: Verify controller creation
- `TestNewChatControllerWithLogger`: Verify logger option injection
- `TestChatControllerMultipleOptions`: Verify multiple options can be chained
- `TestWithLoggerOption`: Verify option application
- `BenchmarkControllerCreation`: Performance benchmark

**Run**: `go test -v ./internal/controller/chat -run Test`

### 2. `controller_integration_test.go` - Integration Tests
**Location**: `internal/controller/chat/controller_integration_test.go`

Tests WebSocket communication using `httptest.Server` and `websocket.Dial`:
- `TestWebSocketEchoHandler`: Full send/receive cycle with echo handler
- `TestWebSocketConnectionClose`: Test connection closure handling
- `TestWebSocketBinaryFrames`: Test binary message frames

**Run**: `go test -v ./internal/controller/chat -run TestWebSocket`

### 3. `mocks_test.go` - Mock Implementations
**Location**: `internal/controller/chat/mocks_test.go`

Provides reusable mock objects for testing:
- `MockWebSocketConn`: Mock WebSocket connection
  - `SendText(msg string)`: Send text messages
  - `SendBinary(data []byte)`: Send binary data
  - `GetMessages()`: Retrieve all sent text messages
  - `GetBinaryMessages()`: Retrieve all sent binary data
  - `Close()`: Close the connection
  - `IsClosed()`: Check connection status
  
- `MockClientManager`: Manager for testing broadcast patterns
  - `Add(id string, conn *MockWebSocketConn)`: Add client
  - `Remove(id string)`: Remove client
  - `Broadcast(msg string)`: Send to all clients
  - `BroadcastExcept(exceptID, msg string)`: Send except one
  - `GetClientCount()`: Get number of connected clients
  - `GetClient(id string)`: Retrieve specific client

### 4. `mocks_usage_test.go` - Mock Usage Examples
**Location**: `internal/controller/chat/mocks_usage_test.go`

Demonstrates using the mocks for testing broadcast scenarios:
- `TestMockWebSocketConn`: Basic mock connection testing
- `TestMockWebSocketConnClosed`: Test closed connection behavior
- `TestMockClientManagerBroadcast`: Test broadcast to multiple clients
- `TestMockClientManagerBroadcastExcept`: Test selective broadcast
- `TestMockClientManagerRemoveClient`: Test client removal
- `TestMockClientManagerGetClient`: Test client retrieval

## Usage Patterns

### Pattern 1: Unit Test with Logger Injection
```go
func TestMyController(t *testing.T) {
    logger, _ := zap.NewDevelopment()
    defer logger.Sync()
    
    ctrl := NewChatController(WithLogger(logger))
    // Test controller behavior
}
```

### Pattern 2: Integration Test with httptest
```go
func TestWebSocketIntegration(t *testing.T) {
    handler := func(ws *websocket.Conn) {
        var msg string
        websocket.Message.Receive(ws, &msg)
        websocket.Message.Send(ws, "response: "+msg)
    }
    
    server := httptest.NewServer(websocket.Server{Handler: handler})
    defer server.Close()
    
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    ws, _ := websocket.Dial(wsURL, "", server.URL)
    // Test communication
}
```

### Pattern 3: Mock-Based Broadcast Testing
```go
func TestBroadcast(t *testing.T) {
    mgr := NewMockClientManager()
    client := NewMockWebSocketConn("user-1")
    mgr.Add("user-1", client)
    
    mgr.Broadcast("message")
    
    if client.GetMessages()[0] != "message" {
        t.Error("broadcast failed")
    }
}
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./internal/controller/chat

# Run specific test
go test -run TestWebSocketEchoHandler ./internal/controller/chat

# Run with verbose output
go test -v ./internal/controller/chat

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Current Test Coverage
- **Unit Tests**: Controller initialization and options (7 tests)
- **Integration Tests**: WebSocket communication (3 tests)
- **Mock Tests**: Client manager and broadcast logic (6 tests)
- **Total**: 16 tests, ~40% statement coverage

## Key Testing Principles

1. **Mock Logger**: Always provide a logger to avoid nil pointer issues
2. **Thread Safety**: MockWebSocketConn and MockClientManager use mutexes
3. **Connection Lifecycle**: Test both open and closed states
4. **Message Types**: Test both text and binary frames
5. **Broadcast Patterns**: Test all-broadcast and selective-broadcast scenarios

## Next Steps for Expanding Tests

1. **Add timeout testing**: Test connection timeouts and cleanup
2. **Add error cases**: Test malformed messages and invalid operations
3. **Add concurrent tests**: Test multiple simultaneous connections
4. **Add protocol tests**: Test message format validation
5. **Add integration tests**: Test with actual app initialization
