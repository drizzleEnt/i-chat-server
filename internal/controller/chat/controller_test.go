package chatctrl

import (
	"testing"

	"go.uber.org/zap"
)

// TestNewChatControllerWithoutOptions verifies controller can be created without options
func TestNewChatControllerWithoutOptions(t *testing.T) {
	ctrl := NewChatController()
	if ctrl == nil {
		t.Fatal("ChatController should not be nil")
	}
}

// TestNewChatControllerWithLogger verifies logger option is properly applied
func TestNewChatControllerWithLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := NewChatController(WithLogger(logger))

	// Type assertion to verify implementation
	impl, ok := ctrl.(*implementation)
	if !ok {
		t.Fatal("Could not cast to implementation")
	}

	if impl.log != logger {
		t.Error("Logger option was not properly applied")
	}
}

// TestChatControllerMultipleOptions verifies multiple options can be chained
func TestChatControllerMultipleOptions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Test multiple options
	ctrl := NewChatController(
		WithLogger(logger),
		WithLogger(logger), // Apply same option twice
	)

	impl, ok := ctrl.(*implementation)
	if !ok {
		t.Fatal("Could not cast to implementation")
	}

	if impl.log != logger {
		t.Error("Logger was not properly set with multiple options")
	}
}

// TestWithLoggerOption verifies the WithLogger option works independently
func TestWithLoggerOption(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	opt := WithLogger(logger)
	impl := &implementation{}
	opt(impl)

	if impl.log != logger {
		t.Error("WithLogger option did not set logger correctly")
	}
}

// Example: BenchmarkControllerCreation can be used to measure performance
func BenchmarkControllerCreation(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewChatController(WithLogger(logger))
	}
}
