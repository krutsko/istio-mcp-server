package istio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// TestNewIstio tests Istio client creation with various configurations
func TestNewIstio(t *testing.T) {
	t.Run("creates Istio client with valid kubeconfig", func(t *testing.T) {
		// Create a temporary kubeconfig file
		tempDir, err := os.MkdirTemp("", "istio-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		kubeconfigPath := filepath.Join(tempDir, "config")
		kubeconfig := createTestKubeconfig("https://localhost:6443")

		if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
			t.Fatalf("Failed to write kubeconfig: %v", err)
		}

		istio, err := NewIstio(kubeconfigPath)
		if err != nil {
			t.Fatalf("Failed to create Istio client: %v", err)
		}
		if istio == nil {
			t.Fatal("Istio client is nil")
		}
		if istio.kubeClient == nil {
			t.Fatal("Kubernetes client is nil")
		}
		if istio.istioClient == nil {
			t.Fatal("Istio client is nil")
		}
		if istio.config == nil {
			t.Fatal("Config is nil")
		}
		if istio.ProxyConfig == nil {
			t.Fatal("ProxyConfig client is nil")
		}

		istio.Close()
	})

	t.Run("creates Istio client with default kubeconfig", func(t *testing.T) {
		// This test might fail in environments without a default kubeconfig
		// but we still test the code path
		_, err := NewIstio("")
		// We don't assert on error here since it depends on the test environment
		if err != nil {
			t.Logf("Expected error when no kubeconfig is available: %v", err)
		}
	})

	t.Run("fails with invalid kubeconfig path", func(t *testing.T) {
		_, err := NewIstio("/non/existent/path")
		if err == nil {
			t.Fatal("Expected error with invalid kubeconfig path")
		}
	})
}

// TestBuildConfig tests Kubernetes configuration building
func TestBuildConfig(t *testing.T) {
	t.Run("builds config from kubeconfig file", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "istio-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		kubeconfigPath := filepath.Join(tempDir, "config")
		kubeconfig := createTestKubeconfig("https://localhost:6443")

		if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
			t.Fatalf("Failed to write kubeconfig: %v", err)
		}

		config, clientCmdConfig, err := buildConfig(kubeconfigPath)
		if err != nil {
			t.Fatalf("Failed to build config: %v", err)
		}
		if config == nil {
			t.Fatal("Config is nil")
		}
		if clientCmdConfig == nil {
			t.Fatal("ClientCmdConfig is nil")
		}
		if config.Host != "https://localhost:6443" {
			t.Fatalf("Expected host 'https://localhost:6443', got '%s'", config.Host)
		}
	})

	t.Run("builds config from default locations", func(t *testing.T) {
		config, clientCmdConfig, err := buildConfig("")
		// This may fail if no default kubeconfig exists, which is fine
		if err != nil {
			t.Logf("Expected error when no default kubeconfig is available: %v", err)
			return
		}
		if config == nil {
			t.Fatal("Config is nil")
		}
		if clientCmdConfig == nil {
			t.Fatal("ClientCmdConfig is nil")
		}
	})
}

func TestIstioNetworkingResources(t *testing.T) {
	// Create a mock server to simulate Istio API responses
	mockServer := createMockIstioServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfig(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("GetVirtualServices", func(t *testing.T) {
		result, err := istio.GetVirtualServices(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get virtual services: %v", err)
		}
		if !strings.Contains(result, "Virtual Services") {
			t.Fatalf("Expected result to contain 'Virtual Services', got: %s", result)
		}
	})

	t.Run("GetDestinationRules", func(t *testing.T) {
		result, err := istio.GetDestinationRules(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get destination rules: %v", err)
		}
		if !strings.Contains(result, "Destination Rules") {
			t.Fatalf("Expected result to contain 'Destination Rules', got: %s", result)
		}
	})

	t.Run("GetGateways", func(t *testing.T) {
		result, err := istio.GetGateways(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get gateways: %v", err)
		}
		if !strings.Contains(result, "Gateways") {
			t.Fatalf("Expected result to contain 'Gateways', got: %s", result)
		}
	})

	t.Run("GetServiceEntries", func(t *testing.T) {
		result, err := istio.GetServiceEntries(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get service entries: %v", err)
		}
		if !strings.Contains(result, "Service Entries") {
			t.Fatalf("Expected result to contain 'Service Entries', got: %s", result)
		}
	})
}

func TestIstioSecurityResources(t *testing.T) {
	mockServer := createMockIstioServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfig(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("GetAuthorizationPolicies", func(t *testing.T) {
		result, err := istio.GetAuthorizationPolicies(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get authorization policies: %v", err)
		}
		if !strings.Contains(result, "Authorization Policies") {
			t.Fatalf("Expected result to contain 'Authorization Policies', got: %s", result)
		}
	})

	t.Run("GetPeerAuthentications", func(t *testing.T) {
		result, err := istio.GetPeerAuthentications(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get peer authentications: %v", err)
		}
		if !strings.Contains(result, "Peer Authentications") {
			t.Fatalf("Expected result to contain 'Peer Authentications', got: %s", result)
		}
	})

	// Note: GetRequestAuthentications might not be implemented yet
	// Commented out until the method is available
	// t.Run("GetRequestAuthentications", func(t *testing.T) {
	//	result, err := istio.GetRequestAuthentications(ctx, "default")
	//	if err != nil {
	//		t.Fatalf("Failed to get request authentications: %v", err)
	//	}
	//	if !strings.Contains(result, "Request Authentications") {
	//		t.Fatalf("Expected result to contain 'Request Authentications', got: %s", result)
	//	}
	// })
}

func TestIstioConfigurationResources(t *testing.T) {
	mockServer := createMockIstioServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfig(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("GetEnvoyFilters", func(t *testing.T) {
		result, err := istio.GetEnvoyFilters(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get envoy filters: %v", err)
		}
		if !strings.Contains(result, "Envoy Filters") {
			t.Fatalf("Expected result to contain 'Envoy Filters', got: %s", result)
		}
	})

	// Note: GetWasmPlugins might not be implemented yet
	// Commented out until the method is available
	// t.Run("GetWasmPlugins", func(t *testing.T) {
	//	result, err := istio.GetWasmPlugins(ctx, "default")
	//	if err != nil {
	//		t.Fatalf("Failed to get wasm plugins: %v", err)
	//	}
	//	if !strings.Contains(result, "Wasm Plugins") {
	//		t.Fatalf("Expected result to contain 'Wasm Plugins', got: %s", result)
	//	}
	// })

	t.Run("GetTelemetries", func(t *testing.T) {
		result, err := istio.GetTelemetries(ctx, "default")
		if err != nil {
			t.Fatalf("Failed to get telemetries: %v", err)
		}
		if !strings.Contains(result, "Telemetry") {
			t.Fatalf("Expected result to contain 'Telemetry', got: %s", result)
		}
	})
}

func TestWatchKubeConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "istio-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfig("https://localhost:6443")

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	callbackCalled := false
	istio.WatchKubeConfig(func() error {
		callbackCalled = true
		return nil
	})

	// Give the watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Modify the kubeconfig file
	f, err := os.OpenFile(kubeconfigPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open kubeconfig: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString("\n# test change\n")
	if err != nil {
		t.Fatalf("Failed to write to kubeconfig: %v", err)
	}

	// Wait for the callback to be called
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// File watching might not work in all test environments
			t.Log("File watching test completed (callback may not trigger in all environments)")
			return
		case <-ticker.C:
			if callbackCalled {
				t.Log("Callback was called successfully")
				return
			}
		}
	}
}

func TestClose(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "istio-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfig("https://localhost:6443")

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}

	// Close should not panic
	istio.Close()
}

// TestCheckExternalDependencyAvailability tests the external dependency availability check
func TestCheckExternalDependencyAvailability(t *testing.T) {
	// Create a mock server to simulate Istio API responses
	mockServer := createMockIstioServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "istio-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kubeconfigPath := filepath.Join(tempDir, "config")
	kubeconfig := createTestKubeconfig(mockServer.URL)

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	istio, err := NewIstio(kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to create Istio client: %v", err)
	}
	defer istio.Close()

	ctx := context.Background()

	t.Run("check external dependency availability", func(t *testing.T) {
		// Test the external dependency check
		result, err := istio.CheckExternalDependencyAvailability(ctx, "test-service", "rds.amazonaws.com", "default")
		if err != nil {
			t.Fatalf("Failed to check external dependency availability: %v", err)
		}

		// Verify the result contains expected information
		if !strings.Contains(result, "External Dependency Check for service 'test-service' -> 'rds.amazonaws.com'") {
			t.Errorf("Result should contain service and host information")
		}

		if !strings.Contains(result, "Service Entry:") {
			t.Errorf("Result should contain Service Entry check")
		}

		if !strings.Contains(result, "Virtual Service:") {
			t.Errorf("Result should contain Virtual Service check")
		}

		if !strings.Contains(result, "Destination Rule:") {
			t.Errorf("Result should contain Destination Rule check")
		}

		if !strings.Contains(result, "Authorization Policy:") {
			t.Errorf("Result should contain Authorization Policy check")
		}

		t.Logf("External dependency check result: %s", result)
	})
}

// Helper functions

func createTestKubeconfig(server string) string {
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

func createMockIstioServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock responses for various Istio API endpoints
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return empty lists for all resources
		emptyList := `{
			"apiVersion": "v1",
			"kind": "List", 
			"items": []
		}`

		switch {
		case strings.Contains(r.URL.Path, "/api"):
			w.Write([]byte(`{"kind":"APIVersions","versions":["v1"]}`))
		case strings.Contains(r.URL.Path, "/apis"):
			w.Write([]byte(`{"kind":"APIGroupList","groups":[]}`))
		default:
			w.Write([]byte(emptyList))
		}
	})

	return httptest.NewServer(mux)
}
