package chatctrl

import (
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

// TestWebSocketEchoHandler demonstrates testing WebSocket communication
// This is an example pattern for testing WS handlers with gorilla/websocket style usage
func TestWebSocketEchoHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create a simple echo WebSocket handler for testing
	echoHandler := func(ws *websocket.Conn) {
		defer ws.Close()
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			t.Logf("Failed to receive message: %v", err)
			return
		}
		// Echo back the message
		err = websocket.Message.Send(ws, "echo: "+msg)
		if err != nil {
			t.Logf("Failed to send message: %v", err)
			return
		}
	}

	// Create test HTTP server
	server := httptest.NewServer(
		websocket.Server{
			Handler: echoHandler,
		},
	)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect WebSocket client
	ws, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer ws.Close()

	// Send a message
	testMessage := "hello websocket"
	_, err = ws.Write([]byte(testMessage))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Receive echo response
	var response string
	err = websocket.Message.Receive(ws, &response)
	if err != nil {
		t.Fatalf("Failed to receive: %v", err)
	}

	// Verify response
	expected := "echo: " + testMessage
	if response != expected {
		t.Errorf("Expected '%s', got '%s'", expected, response)
	}
}

// TestWebSocketConnectionClose demonstrates testing connection closure
func TestWebSocketConnectionClose(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Handler that closes connection immediately
	closeHandler := func(ws *websocket.Conn) {
		logger.Info("client connected", zap.String("remote", ws.RemoteAddr().String()))
		ws.Close()
	}

	server := httptest.NewServer(
		websocket.Server{
			Handler: closeHandler,
		},
	)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer ws.Close()

	// Try to read - should get error since handler closes
	var msg string
	err = websocket.Message.Receive(ws, &msg)
	if err == nil {
		t.Error("Expected error on closed connection, got none")
	}
}

// TestWebSocketBinaryFrames demonstrates testing binary message handling
func TestWebSocketBinaryFrames(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	binaryHandler := func(ws *websocket.Conn) {
		defer ws.Close()
		var data []byte
		err := websocket.Message.Receive(ws, &data)
		if err != nil {
			return
		}
		// Send back same data
		_ = websocket.Message.Send(ws, data)
	}

	server := httptest.NewServer(
		websocket.Server{
			Handler: binaryHandler,
		},
	)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer ws.Close()

	// Send binary data
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	_, err = ws.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Receive response
	var response []byte
	err = websocket.Message.Receive(ws, &response)
	if err != nil {
		t.Fatalf("Failed to receive: %v", err)
	}

	if len(response) != len(testData) {
		t.Errorf("Expected %d bytes, got %d", len(testData), len(response))
	}
}
