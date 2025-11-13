package chatctrl

import (
	"testing"
)

// TestMockWebSocketConn verifies the mock connection works correctly
func TestMockWebSocketConn(t *testing.T) {
	conn := NewMockWebSocketConn("client-1")

	// Test sending messages
	err := conn.SendText("hello")
	if err != nil {
		t.Errorf("SendText failed: %v", err)
	}

	err = conn.SendText("world")
	if err != nil {
		t.Errorf("SendText failed: %v", err)
	}

	// Verify messages were recorded
	messages := conn.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0] != "hello" || messages[1] != "world" {
		t.Errorf("Messages don't match: %v", messages)
	}
}

// TestMockWebSocketConnClosed tests behavior when connection is closed
func TestMockWebSocketConnClosed(t *testing.T) {
	conn := NewMockWebSocketConn("client-1")

	// Close the connection
	err := conn.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Try to send after closing
	err = conn.SendText("should fail")
	if err == nil {
		t.Error("Expected error when sending on closed connection")
	}

	if !conn.IsClosed() {
		t.Error("Connection should be marked as closed")
	}
}

// TestMockClientManagerBroadcast tests broadcasting to multiple clients
func TestMockClientManagerBroadcast(t *testing.T) {
	mgr := NewMockClientManager()

	// Create multiple mock clients
	client1 := NewMockWebSocketConn("user-1")
	client2 := NewMockWebSocketConn("user-2")
	client3 := NewMockWebSocketConn("user-3")

	mgr.Add("user-1", client1)
	mgr.Add("user-2", client2)
	mgr.Add("user-3", client3)

	if mgr.GetClientCount() != 3 {
		t.Errorf("Expected 3 clients, got %d", mgr.GetClientCount())
	}

	// Broadcast a message
	broadcastMsg := "message from user-1"
	count := mgr.Broadcast(broadcastMsg)

	if count != 3 {
		t.Errorf("Expected broadcast to 3 clients, got %d", count)
	}

	// Verify all clients received the message
	if len(client1.GetMessages()) != 1 || client1.GetMessages()[0] != broadcastMsg {
		t.Error("client1 didn't receive broadcast message")
	}
	if len(client2.GetMessages()) != 1 || client2.GetMessages()[0] != broadcastMsg {
		t.Error("client2 didn't receive broadcast message")
	}
	if len(client3.GetMessages()) != 1 || client3.GetMessages()[0] != broadcastMsg {
		t.Error("client3 didn't receive broadcast message")
	}
}

// TestMockClientManagerBroadcastExcept tests broadcasting to all except one client
func TestMockClientManagerBroadcastExcept(t *testing.T) {
	mgr := NewMockClientManager()

	client1 := NewMockWebSocketConn("user-1")
	client2 := NewMockWebSocketConn("user-2")
	client3 := NewMockWebSocketConn("user-3")

	mgr.Add("user-1", client1)
	mgr.Add("user-2", client2)
	mgr.Add("user-3", client3)

	// Broadcast except to user-1
	msg := "private message"
	count := mgr.BroadcastExcept("user-1", msg)

	if count != 2 {
		t.Errorf("Expected broadcast to 2 clients, got %d", count)
	}

	// Verify user-1 didn't get the message, but others did
	if len(client1.GetMessages()) != 0 {
		t.Error("user-1 should not have received the broadcast")
	}

	if len(client2.GetMessages()) != 1 || client2.GetMessages()[0] != msg {
		t.Error("user-2 should have received the broadcast")
	}

	if len(client3.GetMessages()) != 1 || client3.GetMessages()[0] != msg {
		t.Error("user-3 should have received the broadcast")
	}
}

// TestMockClientManagerRemoveClient tests removing a client and broadcasting
func TestMockClientManagerRemoveClient(t *testing.T) {
	mgr := NewMockClientManager()

	client1 := NewMockWebSocketConn("user-1")
	client2 := NewMockWebSocketConn("user-2")

	mgr.Add("user-1", client1)
	mgr.Add("user-2", client2)

	if mgr.GetClientCount() != 2 {
		t.Errorf("Expected 2 clients, got %d", mgr.GetClientCount())
	}

	// Remove user-1
	mgr.Remove("user-1")

	if mgr.GetClientCount() != 1 {
		t.Errorf("Expected 1 client after removal, got %d", mgr.GetClientCount())
	}

	// Broadcast should only reach user-2
	count := mgr.Broadcast("test message")
	if count != 1 {
		t.Errorf("Expected broadcast to reach 1 client, got %d", count)
	}
}

// TestMockClientManagerGetClient tests retrieving a specific client
func TestMockClientManagerGetClient(t *testing.T) {
	mgr := NewMockClientManager()

	client := NewMockWebSocketConn("user-1")
	mgr.Add("user-1", client)

	retrieved := mgr.GetClient("user-1")
	if retrieved == nil {
		t.Error("Failed to retrieve client")
	}

	if retrieved.ID != "user-1" {
		t.Errorf("Expected client ID 'user-1', got '%s'", retrieved.ID)
	}

	// Test getting non-existent client
	notFound := mgr.GetClient("user-99")
	if notFound != nil {
		t.Error("Expected nil for non-existent client")
	}
}
