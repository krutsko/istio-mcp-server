package istio

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// ProxyConfigClient handles Envoy proxy configuration retrieval using istioctl
type ProxyConfigClient struct {
	kubeconfig string
	timeout    time.Duration
}

// NewProxyConfigClient creates a new proxy configuration client
func NewProxyConfigClient(kubeconfig string) *ProxyConfigClient {
	return &ProxyConfigClient{
		kubeconfig: kubeconfig,
		timeout:    30 * time.Second,
	}
}

// GetClusters retrieves cluster configuration from a pod's Envoy proxy
func (p *ProxyConfigClient) GetClusters(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-config", "cluster", fmt.Sprintf("%s.%s", podName, namespace), "-o", "json")
}

// GetListeners retrieves listener configuration from a pod's Envoy proxy
func (p *ProxyConfigClient) GetListeners(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-config", "listener", fmt.Sprintf("%s.%s", podName, namespace), "-o", "json")
}

// GetRoutes retrieves route configuration from a pod's Envoy proxy
func (p *ProxyConfigClient) GetRoutes(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-config", "route", fmt.Sprintf("%s.%s", podName, namespace), "-o", "json")
}

// GetEndpoints retrieves endpoint configuration from a pod's Envoy proxy
func (p *ProxyConfigClient) GetEndpoints(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-config", "endpoint", fmt.Sprintf("%s.%s", podName, namespace), "-o", "json")
}

// GetBootstrap retrieves bootstrap configuration from a pod's Envoy proxy
func (p *ProxyConfigClient) GetBootstrap(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-config", "bootstrap", fmt.Sprintf("%s.%s", podName, namespace), "-o", "json")
}

// GetSecret retrieves secret configuration from a pod's Envoy proxy
func (p *ProxyConfigClient) GetSecret(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-config", "secret", fmt.Sprintf("%s.%s", podName, namespace), "-o", "json")
}

// GetConfigDump retrieves full configuration dump from a pod's Envoy proxy
func (p *ProxyConfigClient) GetConfigDump(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-config", "all", fmt.Sprintf("%s.%s", podName, namespace), "-o", "json")
}

// GetProxyStatus retrieves proxy status information for all pods
func (p *ProxyConfigClient) GetProxyStatus(ctx context.Context) (string, error) {
	return p.execIstioctl(ctx, "proxy-status")
}

// GetProxyStatusForPod retrieves proxy status for a specific pod
func (p *ProxyConfigClient) GetProxyStatusForPod(ctx context.Context, namespace, podName string) (string, error) {
	return p.execIstioctl(ctx, "proxy-status", fmt.Sprintf("%s.%s", podName, namespace))
}

// GetAnalyze performs Istio configuration analysis and reports potential issues
func (p *ProxyConfigClient) GetAnalyze(ctx context.Context, namespace string) (string, error) {
	if namespace != "" {
		return p.execIstioctl(ctx, "analyze", "-n", namespace)
	}
	return p.execIstioctl(ctx, "analyze")
}

// execIstioctl executes istioctl commands with proper error handling and timeout
func (p *ProxyConfigClient) execIstioctl(ctx context.Context, args ...string) (string, error) {
	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Build command arguments
	cmdArgs := []string{}
	if p.kubeconfig != "" {
		cmdArgs = append(cmdArgs, "--kubeconfig", p.kubeconfig)
	}
	cmdArgs = append(cmdArgs, args...)

	// Execute istioctl command
	cmd := exec.CommandContext(ctxWithTimeout, "istioctl", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("istioctl command failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// EnvoyAdminClient handles direct access to Envoy's admin API
type EnvoyAdminClient struct {
	httpClient *http.Client
}

// NewEnvoyAdminClient creates a new Envoy admin API client
func NewEnvoyAdminClient() *EnvoyAdminClient {
	return &EnvoyAdminClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyConfigSummary provides a summary of all proxy configurations in a namespace
func (p *ProxyConfigClient) ProxyConfigSummary(ctx context.Context, namespace string) (string, error) {
	// Get proxy status first to identify pods
	statusOutput, err := p.GetProxyStatus(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get proxy status: %w", err)
	}

	result := fmt.Sprintf("Proxy Configuration Summary for namespace '%s':\n\n", namespace)
	result += "Proxy Status:\n"
	result += statusOutput + "\n\n"

	// Parse the status output to extract pod names for the given namespace
	lines := strings.Split(statusOutput, "\n")
	var namespacePods []string

	for _, line := range lines {
		if strings.Contains(line, "."+namespace) && !strings.Contains(line, "NAME") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				podName := fields[0]
				if strings.Contains(podName, "."+namespace) {
					namespacePods = append(namespacePods, podName)
				}
			}
		}
	}

	if len(namespacePods) == 0 {
		result += fmt.Sprintf("No proxy-enabled pods found in namespace '%s'\n", namespace)
		return result, nil
	}

	result += fmt.Sprintf("Found %d proxy-enabled pods in namespace '%s':\n", len(namespacePods), namespace)
	for _, podName := range namespacePods {
		result += fmt.Sprintf("- %s\n", podName)
	}

	return result, nil
}
