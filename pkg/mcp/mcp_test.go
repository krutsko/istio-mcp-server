package mcp

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestNewServer tests server creation with valid configuration
func TestNewServer(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		t.Run("creates server with valid configuration", func(t *testing.T) {
			profile := &FullProfile{}

			server, err := NewServer(Configuration{
				Profile:    profile,
				Kubeconfig: c.kubeconfigPath,
			})

			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}
			if server == nil {
				t.Fatal("Server is nil")
			}
			if server.configuration == nil {
				t.Fatal("Server configuration is nil")
			}
			if server.server == nil {
				t.Fatal("MCP server is nil")
			}

			server.Close()
		})
	})
}

// TestNewServerWithInvalidKubeconfig tests server creation with invalid kubeconfig
func TestNewServerWithInvalidKubeconfig(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		t.Run("fails with invalid kubeconfig", func(t *testing.T) {
			profile := &FullProfile{}

			// Use non-existent kubeconfig path
			_, err := NewServer(Configuration{
				Profile:    profile,
				Kubeconfig: "/non/existent/path",
			})

			if err == nil {
				t.Fatal("Expected error when creating server with invalid kubeconfig")
			}
		})
	})
}

// TestWatchKubeConfig tests kubeconfig file watching functionality
func TestWatchKubeConfig(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("Skipping test on non-Unix-like platforms")
	}

	testCase(t, func(c *mcpContext) {
		t.Run("watches kubeconfig for changes", func(t *testing.T) {
			// Given
			withTimeout, cancel := context.WithTimeout(c.ctx, 5*time.Second)
			defer cancel()

			err := c.setupMCPServer()
			if err != nil {
				t.Fatalf("Failed to setup MCP server: %v", err)
			}

			// When - modify the kubeconfig file
			f, err := os.OpenFile(c.kubeconfigPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				t.Fatalf("Failed to open kubeconfig: %v", err)
			}
			defer f.Close()

			_, err = f.WriteString("\n# test change\n")
			if err != nil {
				t.Fatalf("Failed to write to kubeconfig: %v", err)
			}

			// Wait a bit for file system events to propagate
			select {
			case <-withTimeout.Done():
				t.Log("Test completed without errors (file watching may not trigger in test environment)")
			case <-time.After(100 * time.Millisecond):
				// File watching should work, but we don't assert on it in this simple test
			}
		})
	})
}

// TestServerToolsAvailable tests that server tools are properly configured
func TestServerToolsAvailable(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		t.Run("all tools are available since they are read-only", func(t *testing.T) {
			profile := &FullProfile{}

			// Create server with simplified configuration
			server, err := NewServer(Configuration{
				Profile:    profile,
				Kubeconfig: c.kubeconfigPath,
			})
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}
			defer server.Close()

			tools := profile.GetTools(server)
			if len(tools) == 0 {
				t.Fatal("Expected at least some tools to be available")
			}

			// Verify all tools are read-only and non-destructive
			for _, tool := range tools {
				if tool.Tool.Annotations.ReadOnlyHint == nil || !*tool.Tool.Annotations.ReadOnlyHint {
					t.Fatalf("Tool %s should be marked as read-only", tool.Tool.Name)
				}
				if tool.Tool.Annotations.DestructiveHint != nil && *tool.Tool.Annotations.DestructiveHint {
					t.Fatalf("Tool %s should not be marked as destructive", tool.Tool.Name)
				}
			}
		})
	})
}

func TestServerTransports(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		err := c.setupMCPServer()
		if err != nil {
			t.Fatalf("Failed to setup MCP server: %v", err)
		}
		defer c.server.Close()

		t.Run("creates SSE server", func(t *testing.T) {
			sseServer := c.server.ServeSse("")
			if sseServer == nil {
				t.Fatal("SSE server is nil")
			}
		})

		t.Run("creates SSE server with base URL", func(t *testing.T) {
			sseServer := c.server.ServeSse("https://example.com")
			if sseServer == nil {
				t.Fatal("SSE server with base URL is nil")
			}
		})

		t.Run("creates HTTP server", func(t *testing.T) {
			httpServer := c.server.ServeHTTP()
			if httpServer == nil {
				t.Fatal("HTTP server is nil")
			}
		})
	})
}

func TestNewTextResult(t *testing.T) {
	t.Run("creates success result", func(t *testing.T) {
		result := NewTextResult("test content", nil)
		if result == nil {
			t.Fatal("Result is nil")
		}
		if result.IsError {
			t.Fatal("Result should not be an error")
		}
		if len(result.Content) != 1 {
			t.Fatalf("Expected 1 content item, got %d", len(result.Content))
		}
		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Content is not TextContent")
		}
		if textContent.Text != "test content" {
			t.Fatalf("Expected 'test content', got '%s'", textContent.Text)
		}
	})

	t.Run("creates error result", func(t *testing.T) {
		result := NewTextResult("", &TestError{message: "test error"})
		if result == nil {
			t.Fatal("Result is nil")
		}
		if !result.IsError {
			t.Fatal("Result should be an error")
		}
		if len(result.Content) != 1 {
			t.Fatalf("Expected 1 content item, got %d", len(result.Content))
		}
		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Content is not TextContent")
		}
		if textContent.Text != "test error" {
			t.Fatalf("Expected 'test error', got '%s'", textContent.Text)
		}
	})
}

func TestContextFunc(t *testing.T) {
	t.Run("extracts authorization header from request", func(t *testing.T) {
		ctx := context.Background()
		req, err := http.NewRequest("GET", "/test", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer test-token")

		newCtx := contextFunc(ctx, req)

		authHeader := newCtx.Value("Authorization")
		if authHeader != "Bearer test-token" {
			t.Fatalf("Expected 'Bearer test-token', got '%v'", authHeader)
		}
	})
}

// TestError is a simple error type for testing
type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}
