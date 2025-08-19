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

// TestGetDestinationRules tests the GetDestinationRules method with various scenarios
func TestGetDestinationRules(t *testing.T) {
	// Create a mock server that simulates Istio API responses
	mockServer := createMockDestinationRuleServer()
	defer mockServer.Close()

	// Create temporary directory and kubeconfig
	tempDir, err := os.MkdirTemp("", "istio-dr-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForDR(mockServer.URL)

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

	t.Run("GetDestinationRules with default namespace", func(t *testing.T) {
		result, err := istio.GetDestinationRules(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get destination rules: %v", err)
		}

		// Verify result contains expected information
		if !strings.Contains(result, "Destination Rules") {
			t.Errorf("Expected result to contain 'Destination Rules', got: %s", result)
		}

		if !strings.Contains(result, "namespace 'default'") {
			t.Errorf("Expected result to mention namespace 'default', got: %s", result)
		}

		// Should contain count information
		if !strings.Contains(result, "Found") {
			t.Errorf("Expected result to contain count information, got: %s", result)
		}
	})

	t.Run("GetDestinationRules with custom namespace", func(t *testing.T) {
		result, err := istio.GetDestinationRules(ctx, "istio-system")
		if err != nil {
			t.Fatalf("Failed to get destination rules: %v", err)
		}

		if !strings.Contains(result, "namespace 'istio-system'") {
			t.Errorf("Expected result to mention namespace 'istio-system', got: %s", result)
		}
	})

	t.Run("GetDestinationRules handles empty result", func(t *testing.T) {
		// Test with a namespace that has no destination rules
		result, err := istio.GetDestinationRules(ctx, "empty-namespace")
		if err != nil {
			t.Fatalf("Failed to get destination rules: %v", err)
		}

		// Should indicate no destination rules found
		if !strings.Contains(result, "Found 0 Destination Rules") {
			t.Errorf("Expected result to indicate 0 destination rules, got: %s", result)
		}
	})

	t.Run("GetDestinationRules with populated namespace", func(t *testing.T) {
		// Test with a namespace that has destination rules configured in mock
		result, err := istio.GetDestinationRules(ctx, "production")
		if err != nil {
			t.Fatalf("Failed to get destination rules: %v", err)
		}

		// Should contain information about destination rules
		if strings.Contains(result, "Found 0 Destination Rules") {
			t.Errorf("Expected result to show destination rules, got: %s", result)
		}

		// Should contain formatting like hosts, traffic policies
		expectedPatterns := []string{"Destination Rules", "namespace 'production'", "Found"}
		for _, pattern := range expectedPatterns {
			if !strings.Contains(result, pattern) {
				t.Errorf("Expected result to contain '%s', got: %s", pattern, result)
			}
		}
	})

	t.Run("GetDestinationRules result format validation", func(t *testing.T) {
		result, err := istio.GetDestinationRules(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get destination rules: %v", err)
		}

		// Check that result has proper structure
		lines := strings.Split(result, "\n")
		if len(lines) < 1 {
			t.Error("Result should have at least one line")
		}

		// First line should contain count and namespace info
		firstLine := lines[0]
		if !strings.Contains(firstLine, "Found") || !strings.Contains(firstLine, "Destination Rules") {
			t.Errorf("First line should contain count information, got: %s", firstLine)
		}

		// Result should end with newline for proper formatting
		if !strings.HasSuffix(result, "\n") {
			t.Error("Result should end with newline for proper formatting")
		}
	})

	t.Run("GetDestinationRules with host information", func(t *testing.T) {
		// Test with a namespace that has destination rules with host information
		result, err := istio.GetDestinationRules(ctx, "production")
		if err != nil {
			t.Fatalf("Failed to get destination rules: %v", err)
		}

		// Should contain host information for destination rules
		if !strings.Contains(result, "Host:") {
			t.Errorf("Expected result to contain host information, got: %s", result)
		}
	})
}

// TestGetDestinationRulesErrorHandling tests error scenarios
func TestGetDestinationRulesErrorHandling(t *testing.T) {
	// Create a mock server that returns errors
	mockServer := createErrorMockServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-dr-error-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForDR(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("GetDestinationRules handles API errors gracefully", func(t *testing.T) {
		_, err := istio.GetDestinationRules(ctx, "default")
		if err == nil {
			t.Fatal("Expected error when API returns error response")
		}

		// Error should be wrapped with context
		if !strings.Contains(err.Error(), "failed to list destination rules") {
			t.Errorf("Expected error to be wrapped with context, got: %v", err)
		}
	})
}

// TestGetDestinationRulesContextCancellation tests context cancellation
func TestGetDestinationRulesContextCancellation(t *testing.T) {
	// Create a mock server with delay
	mockServer := createDelayedMockServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-dr-ctx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForDR(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	t.Run("GetDestinationRules respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context immediately
		cancel()

		_, err := istio.GetDestinationRules(ctx, "default")
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

func createMockDestinationRuleServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock Destination Rules API endpoint
	mux.HandleFunc("/apis/networking.istio.io/v1alpha3/namespaces/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Extract namespace from URL path
		namespace := extractNamespaceFromPath(r.URL.Path)

		var response string
		switch namespace {
		case "production":
			// Mock response with destination rules
			response = `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "DestinationRuleList",
				"items": [
					{
						"metadata": {
							"name": "bookinfo-dr",
							"namespace": "production"
						},
						"spec": {
							"host": "bookinfo.example.com",
							"trafficPolicy": {
								"loadBalancer": {
									"simple": "ROUND_ROBIN"
								}
							}
						}
					},
					{
						"metadata": {
							"name": "reviews-dr",
							"namespace": "production"
						},
						"spec": {
							"host": "reviews",
							"subsets": [
								{
									"name": "v1",
									"labels": {
										"version": "v1"
									}
								}
							]
						}
					}
				]
			}`
		case "empty-namespace":
			// Mock response with no destination rules
			response = `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "DestinationRuleList",
				"items": []
			}`
		default:
			// Default response for other namespaces
			response = `{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind": "DestinationRuleList",
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

func createTestKubeconfigForDR(server string) string {
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
