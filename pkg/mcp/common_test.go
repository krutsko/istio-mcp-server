package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// mcpContext provides common test context and utilities for MCP server tests
type mcpContext struct {
	ctx            context.Context
	server         *Server
	tempDir        string
	kubeconfigPath string
	before         func(c *mcpContext)
}

// MockServer provides a mock Kubernetes API server for testing
type MockServer struct {
	server *httptest.Server
	config string
}

// NewMockServer creates a new mock Kubernetes API server for testing
func NewMockServer() *MockServer {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	config := fmt.Sprintf(`
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: %s
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
`, server.URL)

	ms := &MockServer{
		server: server,
		config: config,
	}

	// Set up default handlers for common Kubernetes API endpoints
	ms.Handle(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"kind":"APIVersions","versions":["v1"],"serverAddressByClientCIDRs":[{"clientCIDR":"0.0.0.0/0"}]}`))
		case "/apis":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"kind":"APIGroupList","apiVersion":"v1","groups":[]}`))
		case "/api/v1":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"kind":"APIResourceList","apiVersion":"v1","resources":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return ms
}

// Handle adds a handler to the mock server
func (ms *MockServer) Handle(handler http.Handler) {
	ms.server.Config.Handler = handler
}

// Close shuts down the mock server
func (ms *MockServer) Close() {
	ms.server.Close()
}

// testCase runs a test with a configured mcpContext
func testCase(t *testing.T, test func(c *mcpContext)) {
	testCaseWithContext(t, &mcpContext{}, test)
}

// testCaseWithContext runs a test with a pre-configured mcpContext
func testCaseWithContext(t *testing.T, ctx *mcpContext, test func(c *mcpContext)) {
	tempDir, err := os.MkdirTemp("", "istio-mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ctx.tempDir = tempDir
	ctx.ctx = context.Background()
	ctx.kubeconfigPath = filepath.Join(tempDir, "config")

	// Create a default kubeconfig if none exists
	if ctx.before == nil {
		// Set up a default kubeconfig
		defaultConfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
`
		if err := os.WriteFile(ctx.kubeconfigPath, []byte(defaultConfig), 0644); err != nil {
			t.Fatalf("Failed to write default kubeconfig: %v", err)
		}
	}

	if ctx.before != nil {
		ctx.before(ctx)
	}

	test(ctx)
}

// withKubeConfig sets up the test context with a specific kubeconfig
func (c *mcpContext) withKubeConfig(config string) {
	if err := os.WriteFile(c.kubeconfigPath, []byte(config), 0644); err != nil {
		panic(fmt.Sprintf("Failed to write kubeconfig: %v", err))
	}
}

// setupMCPServer creates and configures an MCP server for testing
func (c *mcpContext) setupMCPServer() error {
	profile := &FullProfile{}

	server, err := NewServer(Configuration{
		Profile:    profile,
		Kubeconfig: c.kubeconfigPath,
	})
	if err != nil {
		return err
	}

	c.server = server
	return nil
}

// callTool calls a tool on the MCP server and returns the result
func (c *mcpContext) callTool(toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	if c.server == nil {
		if err := c.setupMCPServer(); err != nil {
			return nil, err
		}
	}

	// Find the tool and call it directly
	for _, tool := range c.server.configuration.Profile.GetTools(c.server) {
		if tool.Tool.Name == toolName {
			request := mcp.CallToolRequest{}
			request.Params.Name = toolName
			request.Params.Arguments = args
			return tool.Handler(c.ctx, request)
		}
	}

	return nil, fmt.Errorf("tool %s not found", toolName)
}

// createKubeconfig creates a kubeconfig with the specified configuration
func createKubeconfig(clusters map[string]*api.Cluster, authInfos map[string]*api.AuthInfo, contexts map[string]*api.Context, currentContext string) string {
	config := api.NewConfig()
	config.Clusters = clusters
	config.AuthInfos = authInfos
	config.Contexts = contexts
	config.CurrentContext = currentContext

	bytes, err := clientcmd.Write(*config)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
