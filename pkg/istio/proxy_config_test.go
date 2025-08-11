package istio

import (
	"context"
	"testing"
	"time"
)

// TestProxyConfigClient tests proxy configuration client creation and properties
func TestProxyConfigClient(t *testing.T) {
	// Create a proxy config client
	client := NewProxyConfigClient("")

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.timeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", client.timeout)
	}
}

// TestEnvoyAdminClient tests Envoy admin client creation and properties
func TestEnvoyAdminClient(t *testing.T) {
	// Create an Envoy admin client
	client := NewEnvoyAdminClient()

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", client.httpClient.Timeout)
	}
}

// TestProxyConfigClientIntegration tests proxy configuration client with actual Istio installation
func TestProxyConfigClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewProxyConfigClient("")
	ctx := context.Background()

	// This would work if istioctl is installed and there are pods
	// For now, we'll just test that the method doesn't panic
	_, err := client.GetProxyStatus(ctx)
	// We expect an error since istioctl might not be available
	if err == nil {
		t.Log("GetProxyStatus succeeded - istioctl is available")
	} else {
		t.Logf("GetProxyStatus failed as expected: %v", err)
	}
}
