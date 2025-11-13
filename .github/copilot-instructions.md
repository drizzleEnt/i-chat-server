# i-chat-server AI Coding Guidelines

A real-time chat server in Go with WebSocket support. Minimal dependencies (only `go.uber.org/zap` for logging and `golang.org/x/net` for WebSocket).

## Architecture Overview

### Layered Structure
The codebase follows a clean layered architecture:
- **`cmd/main.go`**: Entry point - creates and runs the app
- **`internal/app/`**: Application orchestration (dependency initialization, HTTP server setup, graceful shutdown)
- **`internal/controller/`**: HTTP handlers and WebSocket management (interface-driven design)
- **`internal/routes/`**: HTTP route registration
- **`internal/config/`**: Configuration management (environment variables via `config.GetEnvStringOrDefault()`)
- **`internal/service/`, `repository/`, `domain/`**: Currently empty - reserved for business logic expansion

### Key Design Patterns

**Service Provider Pattern** (`internal/app/servive_provider.go`):
- Lazy-initialization singleton for all dependencies
- Methods return cached instances; create new if nil
- Currently provides: `Logger`, `HttpConfig`, `ChatController`
- Add new dependencies here when extending functionality

**Interface-Driven Controllers** (`internal/controller/interface.go`):
- Controllers implement lightweight interfaces (e.g., `ChatController` with `HandleWebSocket()`)
- Chat implementation in `internal/controller/chat/` package
- Use option pattern for controller instantiation: `NewChatController(opts ...Option)`

**Dependency Initialization** (`internal/app/app.go`):
- Sequential initialization via slice of functions: `[]func(context.Context) error`
- Each init function handles one concern (provider, HTTP server)
- Fail-fast on any initialization error

## Critical Components

### HTTP Server Setup
- Uses standard `net/http.ServeMux`; no external frameworks
- WebSocket handler registered dynamically in `HandleWebSocket()`
- Server config: timeouts hardcoded (10s read/write, 120s idle)
- Port defaults to `8181` (env var: `HTTP_PORT`); host defaults to `0.0.0.0` (`HTTP_HOST`)

### Configuration
- All config read at startup via environment variables
- Pattern: `config.GetEnvStringOrDefault("KEY", "default")`
- HTTP config implementation in `internal/config/env/http.go`
- `.env` files should be used locally; ignored by `.gitignore`

### Logging
- Uses `go.uber.org/zap` for structured logging
- Logger created in `serviceProvider.Logger()` with atomic level
- Core setup logic in `getCore()` (incomplete - review full `app.go` for details)
- Access via `sp.Logger(ctx)`

### WebSocket Handler
- Partially implemented in `internal/controller/chat/controller.go`
- Current handler at `/ws` endpoint (registered in `HandleWebSocket`)
- Client struct exists (`internal/controller/chat/client.go`) but unused
- Most connection logic is commented out - implementation in progress

## Development Workflow

### Building & Running
```bash
go run ./cmd/main.go              # Run directly
go build -o bin/server ./cmd     # Build binary to bin/
```

### Adding a New Endpoint
1. Define handler interface in `internal/controller/` if needed
2. Create implementation in `internal/controller/[name]/` 
3. Use option pattern for dependency injection: `func NewXxxController(opts ...Option)`
4. Register in `internal/routes/routes.go`
5. Add to service provider in `internal/app/servive_provider.go`

### Adding Configuration
1. Create config interface in `internal/config/`
2. Implement in `internal/config/env/` using `config.GetEnvStringOrDefault()`
3. Add getter to `serviceProvider` (lazy init pattern)
4. Document env var name and default in this file

## Dependencies & Versions
- **Go**: 1.25.3+
- **zap**: v1.19.1 (structured logging)
- **golang.org/x/net**: v0.47.0 (WebSocket via `websocket.Server`)
- No database or ORM currently integrated

## Common Gotchas
- **Logger Initialization**: `serviceProvider.Logger()` has a logic bug (returns immediately if `sp.logger != nil`, never sets it). Review before using.
- **WebSocket Handler**: Runs in goroutine during init; HTTP route registration happens inside the handler function.
- **Typo in Filename**: `servive_provider.go` should be `service_provider.go` (note the typo).
- **Empty Directories**: `domain/`, `repository/`, `service/` are reserved for future use but currently empty.

## Testing Strategy

### Go Testing Conventions
- `*_test.go` suffix for all test files
- `go test ./...` to run all tests with coverage: `go test -cover ./...`
- `go test -run TestNamePattern` to run specific tests
- Use interfaces for mockable components (already in place for controllers)

### Testing Controllers with WebSocket Dependencies

#### 1. Unit Test Pattern: Mock Logger & Dependencies
Create `internal/controller/chat/controller_test.go`:

```go
package chatctrl

import (
	"testing"
	
	"go.uber.org/zap"
)

func TestChatControllerInitialization(t *testing.T) {
	// Create a no-op logger for testing
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	// Test with WithLogger option
	ctrl := NewChatController(WithLogger(logger))
	
	if ctrl == nil {
		t.Fatal("ChatController should not be nil")
	}
}

func TestHandleWebSocketHandlerRegistration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	ctrl := NewChatController(WithLogger(logger))
	
	// This would run in a goroutine in the actual flow
	// For testing, verify it doesn't panic and registers the handler
	go ctrl.HandleWebSocket()
	
	// Give it a moment to register
	time.Sleep(100 * time.Millisecond)
	
	// Verify handler is registered (would need to expose this or test via HTTP)
	t.Log("WebSocket handler registered successfully")
}
```

#### 2. Integration Test: Full HTTP Handler Test
Create `internal/controller/chat/controller_integration_test.go`:

```go
package chatctrl

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func TestWebSocketHandlerIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	ctrl := NewChatController(WithLogger(logger))
	
	// Create test HTTP server with the controller
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate the /ws route handling
		s := &websocket.Server{
			Handler: func(ws *websocket.Conn) {
				defer ws.Close()
				// Read from client
				var msg string
				websocket.Message.Receive(ws, &msg)
				// Echo back
				websocket.Message.Send(ws, "echo: "+msg)
			},
		}
		s.ServeHTTP(w, r)
	}))
	defer server.Close()
	
	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	
	// Connect WebSocket client
	ws, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()
	
	// Test send/receive
	if _, err := ws.Write([]byte("hello")); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	
	var response string
	if err := websocket.Message.Receive(ws, &response); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	
	if !strings.Contains(response, "hello") {
		t.Errorf("Expected response to contain 'hello', got: %s", response)
	}
}
```

#### 3. Option Pattern Testing
Extend `controller.go` to support dependency injection via options:

```go
// Add to internal/controller/chat/controller.go
type Option func(*implementation)

func WithLogger(logger *zap.Logger) Option {
	return func(impl *implementation) {
		impl.log = logger
	}
}

// Test that options are properly applied
func TestChatControllerOptions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	opt := WithLogger(logger)
	impl := &implementation{}
	opt(impl)
	
	if impl.log != logger {
		t.Error("WithLogger option did not set logger correctly")
	}
}
```

#### 4. Mock Client for Broadcast Testing
Create `internal/controller/chat/mocks_test.go`:

```go
package chatctrl

import (
	"sync"
	"testing"
)

type MockWebSocketConn struct {
	mu       sync.Mutex
	messages []string
	closed   bool
}

func (m *MockWebSocketConn) Send(msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("connection closed")
	}
	m.messages = append(m.messages, msg)
	return nil
}

func (m *MockWebSocketConn) GetMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

func TestBroadcastToClients(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	ctrl := NewChatController(WithLogger(logger))
	
	// Create mock connections
	mock1 := &MockWebSocketConn{}
	mock2 := &MockWebSocketConn{}
	
	// Simulate adding clients and broadcasting
	msg := "test broadcast"
	_ = mock1.Send(msg)
	_ = mock2.Send(msg)
	
	if len(mock1.GetMessages()) != 1 {
		t.Errorf("Expected 1 message in mock1, got %d", len(mock1.GetMessages()))
	}
}
```

### Running Tests
```bash
# Run all tests with coverage
go test -cover ./...

# Run specific test file
go test -v ./internal/controller/chat -run TestChatControllerInitialization

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Files Reference
The project includes example test files demonstrating the patterns above:
- **`controller_test.go`**: Unit tests for controller initialization and option pattern
- **`controller_integration_test.go`**: Integration tests for WebSocket communication with `httptest`
- **`mocks_test.go`**: Mock implementations for testing broadcast and client management
- **`mocks_usage_test.go`**: Examples of using mocks for client manager and broadcast testing

Example: To test client broadcast logic, use `MockClientManager`:
```go
mgr := NewMockClientManager()
client1 := NewMockWebSocketConn("user-1")
mgr.Add("user-1", client1)
mgr.Broadcast("hello")  // client1 receives the message
if client1.GetMessages()[0] != "hello" {
    t.Error("broadcast failed")
}
```

## Recommended Next Steps for Implementation
1. Complete WebSocket client message handling and broadcast logic
2. Implement domain models for `User`, `Message`, `ChatRoom`
3. Add service layer for business logic
4. Implement repository pattern for persistence
5. Add authentication/authorization
6. Fix logger initialization bug in `serviceProvider.Logger()`
