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

// TestGetServices tests the GetServices method with various scenarios
func TestGetServices(t *testing.T) {
	// Create a mock server that simulates Kubernetes API responses
	mockServer := createMockServicesServer()
	defer mockServer.Close()

	// Create temporary directory and kubeconfig
	tempDir, err := os.MkdirTemp("", "istio-services-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForServices(mockServer.URL)

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

	t.Run("GetServices with default namespace", func(t *testing.T) {
		result, err := istio.GetServices(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Verify result contains expected information
		if !strings.Contains(result, "Services in namespace 'default'") {
			t.Errorf("Expected result to contain 'Services in namespace 'default'', got: %s", result)
		}

		if !strings.Contains(result, "Found") {
			t.Errorf("Expected result to contain count information, got: %s", result)
		}
	})

	t.Run("GetServices with custom namespace", func(t *testing.T) {
		result, err := istio.GetServices(ctx, "production")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		if !strings.Contains(result, "namespace 'production'") {
			t.Errorf("Expected result to mention namespace 'production', got: %s", result)
		}
	})

	t.Run("GetServices handles empty result", func(t *testing.T) {
		// Test with a namespace that has no services
		result, err := istio.GetServices(ctx, "empty-namespace")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Should indicate no services found
		if !strings.Contains(result, "Found 0 services") {
			t.Errorf("Expected result to indicate 0 services, got: %s", result)
		}

		if !strings.Contains(result, "No services found in this namespace") {
			t.Errorf("Expected result to contain 'No services found in this namespace', got: %s", result)
		}
	})

	t.Run("GetServices with populated namespace", func(t *testing.T) {
		// Test with a namespace that has services configured in mock
		result, err := istio.GetServices(ctx, "production")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Should contain information about services
		if strings.Contains(result, "Found 0 services") {
			t.Errorf("Expected result to show services, got: %s", result)
		}

		// Should contain formatting like service types
		expectedPatterns := []string{"Services", "namespace 'production'", "Found"}
		for _, pattern := range expectedPatterns {
			if !strings.Contains(result, pattern) {
				t.Errorf("Expected result to contain '%s', got: %s", pattern, result)
			}
		}
	})

	t.Run("GetServices result format validation", func(t *testing.T) {
		result, err := istio.GetServices(ctx, "production")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Check that result has proper structure
		lines := strings.Split(result, "\n")
		if len(lines) < 1 {
			t.Error("Result should have at least one line")
		}

		// First line should contain namespace info
		firstLine := lines[0]
		if !strings.Contains(firstLine, "Services in namespace") {
			t.Errorf("First line should contain namespace information, got: %s", firstLine)
		}

		// Result should contain next step information
		if !strings.Contains(result, "Next step: Use 'get-pods-by-service'") {
			t.Errorf("Result should contain next step information")
		}
	})

	t.Run("GetServices with different service types", func(t *testing.T) {
		// Test with a namespace that has different types of services
		result, err := istio.GetServices(ctx, "mixed-services")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Should contain different service type sections
		expectedServiceTypes := []string{"ClusterIP Services:", "NodePort Services:", "LoadBalancer Services:", "Headless Services:"}
		for _, serviceType := range expectedServiceTypes {
			if !strings.Contains(result, serviceType) {
				t.Errorf("Expected result to contain service type '%s', got: %s", serviceType, result)
			}
		}
	})

	t.Run("GetServices with ClusterIP services", func(t *testing.T) {
		result, err := istio.GetServices(ctx, "clusterip-namespace")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Should contain ClusterIP service information
		if !strings.Contains(result, "ClusterIP Services:") {
			t.Errorf("Expected result to contain ClusterIP services section")
		}

		if !strings.Contains(result, "(ClusterIP:") {
			t.Errorf("Expected result to contain ClusterIP information")
		}
	})

	t.Run("GetServices with NodePort services", func(t *testing.T) {
		result, err := istio.GetServices(ctx, "nodeport-namespace")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Should contain NodePort service information
		if !strings.Contains(result, "NodePort Services:") {
			t.Errorf("Expected result to contain NodePort services section")
		}

		if !strings.Contains(result, "(NodePort:") {
			t.Errorf("Expected result to contain NodePort information")
		}
	})

	t.Run("GetServices with LoadBalancer services", func(t *testing.T) {
		result, err := istio.GetServices(ctx, "loadbalancer-namespace")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Should contain LoadBalancer service information
		if !strings.Contains(result, "LoadBalancer Services:") {
			t.Errorf("Expected result to contain LoadBalancer services section")
		}

		if !strings.Contains(result, "(LoadBalancer:") {
			t.Errorf("Expected result to contain LoadBalancer information")
		}
	})

	t.Run("GetServices with Headless services", func(t *testing.T) {
		result, err := istio.GetServices(ctx, "headless-namespace")
		if err != nil {
			t.Fatalf("Failed to get services: %v", err)
		}

		// Should contain Headless service information
		if !strings.Contains(result, "Headless Services:") {
			t.Errorf("Expected result to contain Headless services section")
		}

		if !strings.Contains(result, "(Headless)") {
			t.Errorf("Expected result to contain Headless information")
		}
	})
}

// TestGetServicesErrorHandling tests error scenarios
func TestGetServicesErrorHandling(t *testing.T) {
	// Create a mock server that returns errors
	mockServer := createErrorMockServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-services-error-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForServices(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("GetServices handles API errors gracefully", func(t *testing.T) {
		_, err := istio.GetServices(ctx, "default")
		if err == nil {
			t.Fatal("Expected error when API returns error response")
		}

		// Error should be wrapped with context
		if !strings.Contains(err.Error(), "failed to list services") {
			t.Errorf("Expected error to be wrapped with context, got: %v", err)
		}
	})
}

// TestGetServicesContextCancellation tests context cancellation
func TestGetServicesContextCancellation(t *testing.T) {
	// Create a mock server with delay
	mockServer := createDelayedMockServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-services-ctx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfigForServices(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	t.Run("GetServices respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context immediately
		cancel()

		_, err := istio.GetServices(ctx, "default")
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

func createMockServicesServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock Services API endpoint
	mux.HandleFunc("/api/v1/namespaces/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Extract namespace from URL path
		namespace := extractNamespaceFromServicesPath(r.URL.Path)

		var response string
		switch namespace {
		case "production":
			// Mock response with mixed service types
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
				"items": [
					{
						"metadata": {
							"name": "web-service",
							"namespace": "production"
						},
						"spec": {
							"type": "ClusterIP",
							"clusterIP": "10.96.1.1"
						}
					},
					{
						"metadata": {
							"name": "api-service",
							"namespace": "production"
						},
						"spec": {
							"type": "NodePort",
							"clusterIP": "10.96.1.2"
						}
					}
				]
			}`
		case "mixed-services":
			// Mock response with all service types
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
				"items": [
					{
						"metadata": {"name": "clusterip-svc"},
						"spec": {"type": "ClusterIP", "clusterIP": "10.96.1.1"}
					},
					{
						"metadata": {"name": "nodeport-svc"},
						"spec": {"type": "NodePort", "clusterIP": "10.96.1.2"}
					},
					{
						"metadata": {"name": "loadbalancer-svc"},
						"spec": {"type": "LoadBalancer", "clusterIP": "10.96.1.3"},
						"status": {
							"loadBalancer": {
								"ingress": [{"ip": "192.168.1.100"}]
							}
						}
					},
					{
						"metadata": {"name": "headless-svc"},
						"spec": {"type": "ClusterIP", "clusterIP": "None"}
					}
				]
			}`
		case "clusterip-namespace":
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
				"items": [
					{
						"metadata": {"name": "app-service"},
						"spec": {"type": "ClusterIP", "clusterIP": "10.96.1.10"}
					}
				]
			}`
		case "nodeport-namespace":
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
				"items": [
					{
						"metadata": {"name": "external-service"},
						"spec": {"type": "NodePort", "clusterIP": "10.96.1.20"}
					}
				]
			}`
		case "loadbalancer-namespace":
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
				"items": [
					{
						"metadata": {"name": "public-service"},
						"spec": {"type": "LoadBalancer", "clusterIP": "10.96.1.30"},
						"status": {
							"loadBalancer": {
								"ingress": [{"hostname": "lb.example.com"}]
							}
						}
					}
				]
			}`
		case "headless-namespace":
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
				"items": [
					{
						"metadata": {"name": "statefulset-service"},
						"spec": {"type": "ClusterIP", "clusterIP": "None"}
					}
				]
			}`
		case "empty-namespace":
			// Mock response with no services
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
				"items": []
			}`
		default:
			// Default response for other namespaces
			response = `{
				"apiVersion": "v1",
				"kind": "ServiceList",
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

	return httptest.NewServer(mux)
}

func createTestKubeconfigForServices(server string) string {
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

func extractNamespaceFromServicesPath(path string) string {
	// Extract namespace from path like: /api/v1/namespaces/default/services
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "namespaces" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "default"
}
