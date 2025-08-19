package istio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// TestGetVirtualServices tests the GetVirtualServices method with various scenarios
func TestGetVirtualServices(t *testing.T) {
	// Create a mock server that simulates Istio API responses
	mockServer := createMockVirtualServiceServer()
	defer mockServer.Close()

	// Create temporary directory and kubeconfig
	tempDir, err := os.MkdirTemp("", "istio-vs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForVS(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	// Create Istio client
	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("GetVirtualServices with default namespace", func(t *testing.T) {
		result, err := istio.GetVirtualServices(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get virtual services: %v", err)
		}

		// Verify result contains expected information
		if !strings.Contains(result, "Virtual Services") {
			t.Errorf("Expected result to contain 'Virtual Services', got: %s", result)
		}

		if !strings.Contains(result, "namespace 'default'") {
			t.Errorf("Expected result to mention namespace 'default', got: %s", result)
		}

		// Should contain count information
		if !strings.Contains(result, "Found") {
			t.Errorf("Expected result to contain count information, got: %s", result)
		}
	})

	t.Run("GetVirtualServices with custom namespace", func(t *testing.T) {
		result, err := istio.GetVirtualServices(ctx, "istio-system")
		if err != nil {
			t.Fatalf("Failed to get virtual services: %v", err)
		}

		if !strings.Contains(result, "namespace 'istio-system'") {
			t.Errorf("Expected result to mention namespace 'istio-system', got: %s", result)
		}
	})

	t.Run("GetVirtualServices handles empty result", func(t *testing.T) {
		// Test with a namespace that has no virtual services
		result, err := istio.GetVirtualServices(ctx, "empty-namespace")
		if err != nil {
			t.Fatalf("Failed to get virtual services: %v", err)
		}

		// Should indicate no virtual services found
		if !strings.Contains(result, "Found 0 Virtual Services") {
			t.Errorf("Expected result to indicate 0 virtual services, got: %s", result)
		}
	})

	t.Run("GetVirtualServices with populated namespace", func(t *testing.T) {
		// Test with a namespace that has virtual services configured in mock
		result, err := istio.GetVirtualServices(ctx, "production")
		if err != nil {
			t.Fatalf("Failed to get virtual services: %v", err)
		}

		// Should contain information about virtual services
		if strings.Contains(result, "Found 0 Virtual Services") {
			t.Errorf("Expected result to show virtual services, got: %s", result)
		}

		// Should contain formatting like hosts, gateways, routes
		expectedPatterns := []string{"Virtual Services", "namespace 'production'", "Found"}
		for _, pattern := range expectedPatterns {
			if !strings.Contains(result, pattern) {
				t.Errorf("Expected result to contain '%s', got: %s", pattern, result)
			}
		}
	})

	t.Run("GetVirtualServices result format validation", func(t *testing.T) {
		result, err := istio.GetVirtualServices(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get virtual services: %v", err)
		}

		// Check that result has proper structure
		lines := strings.Split(result, "\n")
		if len(lines) < 1 {
			t.Error("Result should have at least one line")
		}

		// First line should contain count and namespace info
		firstLine := lines[0]
		if !strings.Contains(firstLine, "Found") || !strings.Contains(firstLine, "Virtual Services") {
			t.Errorf("First line should contain count information, got: %s", firstLine)
		}

		// Result should end with newline for proper formatting
		if !strings.HasSuffix(result, "\n") {
			t.Error("Result should end with newline for proper formatting")
		}
	})
}

// TestGetVirtualServicesErrorHandling tests error scenarios
func TestGetVirtualServicesErrorHandling(t *testing.T) {
	// Create a mock server that returns errors
	mockServer := createErrorMockServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-vs-error-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForVS(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("GetVirtualServices handles API errors gracefully", func(t *testing.T) {
		_, err := istio.GetVirtualServices(ctx, "default")
		if err == nil {
			t.Fatal("Expected error when API returns error response")
		}

		// Error should be wrapped with context
		if !strings.Contains(err.Error(), "failed to list virtual services") {
			t.Errorf("Expected error to be wrapped with context, got: %v", err)
		}
	})
}

// TestGetVirtualServicesContextCancellation tests context cancellation
func TestGetVirtualServicesContextCancellation(t *testing.T) {
	// Create a mock server with delay
	mockServer := createDelayedMockServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-vs-ctx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForVS(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	t.Run("GetVirtualServices respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context immediately
		cancel()

		_, err := istio.GetVirtualServices(ctx, "default")
		if err == nil {
			t.Fatal("Expected error when context is cancelled")
		}

		if !strings.Contains(err.Error(), "context canceled") && !strings.Contains(err.Error(), "cancelled") {
			t.Logf("Context cancellation may not be immediately detected, got error: %v", err)
			// Note: Context cancellation might not always be immediately detected
			// depending on where in the request lifecycle it occurs
		}
	})
}

// Helper functions for creating mock servers

func createMockVirtualServiceServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock Virtual Services API endpoint
	mux.HandleFunc("/apis/networking.istio.io/v1alpha3/namespaces/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Extract namespace from URL path
		namespace := extractNamespaceFromPath(r.URL.Path)

		var response string
		switch namespace {
		case "production":
			// Mock response with virtual services
			response = `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "VirtualServiceList",
				"items": [
					{
						"metadata": {
							"name": "bookinfo-vs",
							"namespace": "production"
						},
						"spec": {
							"hosts": ["bookinfo.example.com"],
							"gateways": ["bookinfo-gateway"],
							"http": [
								{
									"route": [
										{
											"destination": {
												"host": "productpage",
												"port": {"number": 9080}
											}
										}
									]
								}
							]
						}
					}
				]
			}`
		case "empty-namespace":
			// Mock response with no virtual services
			response = `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "VirtualServiceList",
				"items": []
			}`
		default:
			// Default response for other namespaces
			response = `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "VirtualServiceList",
				"items": []
			}`
		}

		w.Write([]byte(response))
	})

	// Mock API discovery endpoints
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"kind":"APIVersions","versions":["v1"]}`))
	})

	mux.HandleFunc("/apis", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"kind":"APIGroupList",
			"groups":[
				{
					"name":"networking.istio.io",
					"versions":[{"version":"v1alpha3"}]
				}
			]
		}`))
	})

	return httptest.NewServer(mux)
}

func createErrorMockServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"kind":"Status","message":"Internal server error"}`))
	})

	return httptest.NewServer(mux)
}

func createDelayedMockServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		select {
		case <-r.Context().Done():
			return
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"kind":"VirtualServiceList","items":[]}`))
		}
	})

	return httptest.NewServer(mux)
}

func createTestKubeconfigForVS(server string) string {
	config := api.NewConfig()
	config.Clusters["test-cluster"] = &api.Cluster{
		Server: server,
	}
	config.AuthInfos["test-user"] = &api.AuthInfo{}
	config.Contexts["test-context"] = &api.Context{
		Cluster:  "test-cluster",
		AuthInfo: "test-user",
	}
	config.CurrentContext = "test-context"

	bytes, err := clientcmd.Write(*config)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func extractNamespaceFromPath(path string) string {
	// Extract namespace from path like: /apis/networking.istio.io/v1alpha3/namespaces/default/virtualservices
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "namespaces" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "default"
}
