package mcp

import (
	"context"
	"fmt"
	"slices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Profile defines the interface for MCP server profiles
type Profile interface {
	GetName() string
	GetDescription() string
	GetTools(s *Server) []server.ServerTool
}

// Profiles contains all available MCP server profiles
var Profiles = []Profile{
	&FullProfile{},
}

// ProfileNames contains the names of all available profiles
var ProfileNames []string

// ProfileFromString returns a profile by name
func ProfileFromString(name string) Profile {
	for _, profile := range Profiles {
		if profile.GetName() == name {
			return profile
		}
	}
	return nil
}

// FullProfile provides access to all Istio MCP tools
type FullProfile struct{}

// GetName returns the profile name
func (p *FullProfile) GetName() string {
	return "full"
}

// GetDescription returns the profile description
func (p *FullProfile) GetDescription() string {
	return "Complete profile with all Istio service mesh tools for networking, security, configuration, and proxy debugging across multiple namespaces"
}

// GetTools returns all available tools for the full profile
func (p *FullProfile) GetTools(s *Server) []server.ServerTool {
	return slices.Concat(
		s.initNetworkingTools(),
		s.initSecurityTools(),
		s.initConfigurationTools(),
		s.initProxyConfigTools(),
	)
}

// initNetworkingTools initializes networking-related Istio tools
func (s *Server) initNetworkingTools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("get-virtual-services",
				mcp.WithDescription("Get Istio Virtual Services configuration from any namespace. Virtual Services define routing rules for services in the Istio service mesh, including traffic splitting, fault injection, and retry policies. Use this to inspect traffic routing configuration across namespaces."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Istio services can span multiple namespaces."),
				),
				mcp.WithTitleAnnotation("Istio: Virtual Services"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getVirtualServices,
		},
		{
			Tool: mcp.NewTool("get-destination-rules",
				mcp.WithDescription("Get Istio Destination Rules from any namespace. Destination Rules define policies for traffic to services, including load balancing, connection pooling, and outlier detection. Essential for understanding service mesh traffic policies."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Check multiple namespaces for complete Istio configuration."),
				),
				mcp.WithTitleAnnotation("Istio: Destination Rules"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getDestinationRules,
		},
		{
			Tool: mcp.NewTool("get-gateways",
				mcp.WithDescription("Get Istio Gateways from any namespace. Gateways configure load balancers for incoming traffic to the service mesh. Use this to inspect ingress/egress configuration and external access patterns."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Gateway configurations may exist in ingress or dedicated namespaces."),
				),
				mcp.WithTitleAnnotation("Istio: Gateways"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getGateways,
		},
		{
			Tool: mcp.NewTool("get-service-entries",
				mcp.WithDescription("Get Istio Service Entries from any namespace. Service Entries allow adding external services to the service mesh registry. Use this to inspect external service configurations and mesh expansion settings."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). External service configurations may be centralized in specific namespaces."),
				),
				mcp.WithTitleAnnotation("Istio: Service Entries"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getServiceEntries,
		},
	}
}

// initSecurityTools initializes security-related Istio tools
func (s *Server) initSecurityTools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("get-authorization-policies",
				mcp.WithDescription("Get Istio Authorization Policies from any namespace. Authorization Policies control access to services in the Istio service mesh, defining who can access what resources. Use this to inspect security policies and access control configurations across namespaces."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Security policies may be defined in multiple namespaces for different service boundaries."),
				),
				mcp.WithTitleAnnotation("Istio: Authorization Policies"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getAuthorizationPolicies,
		},
		{
			Tool: mcp.NewTool("get-peer-authentications",
				mcp.WithDescription("Get Istio Peer Authentications from any namespace. Peer Authentication policies define mutual TLS settings and authentication requirements for service-to-service communication. Use this to inspect mTLS configuration and security posture."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Authentication policies may be namespace-specific or inherited from mesh-wide settings."),
				),
				mcp.WithTitleAnnotation("Istio: Peer Authentications"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getPeerAuthentications,
		},
	}
}

// initConfigurationTools initializes configuration-related Istio tools
func (s *Server) initConfigurationTools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("discover-istio-namespaces",
				mcp.WithDescription("Discover namespaces that have pods with Istio sidecars and rank them by injection density. This tool helps identify the most probable best namespace for Istio operations by analyzing which namespaces have the most Istio-injected workloads. Use this to prioritize which namespaces to investigate first for Istio configuration and traffic analysis."),
				mcp.WithTitleAnnotation("Istio: Namespace Discovery"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.discoverIstioNamespaces,
		},
		{
			Tool: mcp.NewTool("get-envoy-filters",
				mcp.WithDescription("Get Istio Envoy Filters from any namespace. Envoy Filters allow custom configuration of Envoy proxy behavior, including custom filters, listeners, and clusters. Use this to inspect advanced Istio service mesh configurations."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Custom Envoy configurations may be applied to specific namespaces or workloads."),
				),
				mcp.WithTitleAnnotation("Istio: Envoy Filters"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getEnvoyFilters,
		},
		{
			Tool: mcp.NewTool("get-telemetry",
				mcp.WithDescription("Get Istio Telemetry configurations from any namespace. Telemetry policies define observability settings including metrics, tracing, and logging for the service mesh. Use this to inspect monitoring and observability configurations."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Telemetry policies may be namespace-specific or inherited from mesh-wide settings."),
				),
				mcp.WithTitleAnnotation("Istio: Telemetry"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getTelemetries,
		},
		{
			Tool: mcp.NewTool("get-istio-config",
				mcp.WithDescription("Get comprehensive Istio configuration summary for any namespace. This provides an overview of all Istio resources including Virtual Services, Destination Rules, Gateways, Security Policies, and more. Use this for complete Istio service mesh configuration analysis."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to query (defaults to 'default'). Provides complete Istio configuration overview for the specified namespace."),
				),
				mcp.WithTitleAnnotation("Istio: Configuration Summary"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getIstioConfigSummary,
		},
		{
			Tool: mcp.NewTool("check-external-dependency-availability",
				mcp.WithDescription("Check if an external dependency (like RDS, S3, etc.) is properly configured and accessible for a specific service. This tool validates that all required Istio resources (Service Entries, Virtual Services, Destination Rules, Authorization Policies) exist and are properly configured to allow the service to access the external dependency."),
				mcp.WithString("service-name",
					mcp.Description("Name of the service that needs to access the external dependency"),
					mcp.Required(),
				),
				mcp.WithString("external-host",
					mcp.Description("External hostname to check (e.g., 'rds.amazonaws.com', 's3.amazonaws.com')"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the service (defaults to 'default'). The tool will check for Istio resources in this namespace and globally."),
				),
				mcp.WithTitleAnnotation("Istio: External Dependency Check"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.checkExternalDependencyAvailability,
		},
		{
			Tool: mcp.NewTool("get-services",
				mcp.WithDescription("List all Kubernetes services in a namespace. This is the first step in the workflow to find pods for proxy commands: 1) Use this tool to discover available services, 2) Then use 'get-pods-by-service' to find the specific pods backing a service, 3) Finally use proxy commands (get-proxy-clusters, get-proxy-status, etc.) with the discovered pod names. Perfect for understanding the service landscape before diving into Istio proxy configuration."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to list services from (defaults to 'default'). Services are the entry points to your applications."),
				),
				mcp.WithTitleAnnotation("Kubernetes: Service Discovery"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getServices,
		},
		{
			Tool: mcp.NewTool("get-pods-by-service",
				mcp.WithDescription("Find all pods backing a specific Kubernetes service - essential for the proxy command workflow. After discovering services with 'get-services', use this tool to find the exact pod names you need for Istio proxy commands. Shows running vs non-running pods, Istio sidecar status, and provides ready-to-use pod names for proxy debugging commands like get-proxy-clusters, get-proxy-status, get-proxy-listeners, etc. This is step 2 in the service→pod→proxy command workflow."),
				mcp.WithString("namespace",
					mcp.Description("Namespace containing the service (defaults to 'default')"),
				),
				mcp.WithString("service",
					mcp.Description("Service name to find backing pods for (use 'get-services' first to discover available services)"),
					mcp.Required(),
				),
				mcp.WithTitleAnnotation("Kubernetes: Service Pod Discovery"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getPodsByService,
		},
	}
}

// initProxyConfigTools initializes proxy configuration-related Istio tools
func (s *Server) initProxyConfigTools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("get-proxy-clusters",
				mcp.WithDescription("Get Envoy cluster configuration from any Istio proxy pod. Clusters represent upstream services and their load balancing settings. Use this for debugging service connectivity and load balancing issues in the Istio service mesh."),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the pod (defaults to 'default'). Istio proxies can be in any namespace where services are deployed."),
				),
				mcp.WithString("pod",
					mcp.Description("Pod name containing the Istio proxy (sidecar)"),
					mcp.Required(),
				),
				mcp.WithTitleAnnotation("Istio: Proxy Clusters"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getProxyClusters,
		},
		{
			Tool: mcp.NewTool("get-proxy-listeners",
				mcp.WithDescription("Get Envoy listener configuration from any Istio proxy pod. Listeners define how the proxy accepts incoming connections. Use this for debugging network connectivity and port binding issues in the service mesh."),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the pod (defaults to 'default'). Check the namespace where your service pods are deployed."),
				),
				mcp.WithString("pod",
					mcp.Description("Pod name containing the Istio proxy (sidecar)"),
					mcp.Required(),
				),
				mcp.WithTitleAnnotation("Istio: Proxy Listeners"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getProxyListeners,
		},
		{
			Tool: mcp.NewTool("get-proxy-routes",
				mcp.WithDescription("Get Envoy route configuration from any Istio proxy pod. Routes define how requests are matched and routed to clusters. Use this for debugging traffic routing and Virtual Service configuration issues."),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the pod (defaults to 'default'). Route configurations reflect Virtual Service rules applied to the pod."),
				),
				mcp.WithString("pod",
					mcp.Description("Pod name containing the Istio proxy (sidecar)"),
					mcp.Required(),
				),
				mcp.WithTitleAnnotation("Istio: Proxy Routes"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getProxyRoutes,
		},
		{
			Tool: mcp.NewTool("get-proxy-endpoints",
				mcp.WithDescription("Get Envoy endpoint configuration from any Istio proxy pod. Endpoints represent the actual instances of upstream services. Use this for debugging service discovery and endpoint health issues."),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the pod (defaults to 'default'). Endpoint configurations show service discovery results."),
				),
				mcp.WithString("pod",
					mcp.Description("Pod name containing the Istio proxy (sidecar)"),
					mcp.Required(),
				),
				mcp.WithTitleAnnotation("Istio: Proxy Endpoints"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getProxyEndpoints,
		},
		{
			Tool: mcp.NewTool("get-proxy-bootstrap",
				mcp.WithDescription("Get Envoy bootstrap configuration from any Istio proxy pod. Bootstrap config contains the initial proxy configuration including admin interface settings. Use this for debugging proxy startup and configuration issues."),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the pod (defaults to 'default'). Bootstrap config is generated during proxy initialization."),
				),
				mcp.WithString("pod",
					mcp.Description("Pod name containing the Istio proxy (sidecar)"),
					mcp.Required(),
				),
				mcp.WithTitleAnnotation("Istio: Proxy Bootstrap"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getProxyBootstrap,
		},
		{
			Tool: mcp.NewTool("get-proxy-config-dump",
				mcp.WithDescription("Get full Envoy configuration dump from any Istio proxy pod. This provides complete proxy configuration including all listeners, clusters, routes, and endpoints. Use this for comprehensive Istio proxy debugging and troubleshooting."),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the pod (defaults to 'default'). Full config dump shows complete proxy state."),
				),
				mcp.WithString("pod",
					mcp.Description("Pod name containing the Istio proxy (sidecar)"),
					mcp.Required(),
				),
				mcp.WithTitleAnnotation("Istio: Proxy Config Dump"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getProxyConfigDump,
		},
		{
			Tool: mcp.NewTool("get-proxy-status",
				mcp.WithDescription("Get proxy status information for all Istio proxies or a specific pod. Shows proxy sync status, configuration version, and connectivity health. Use this to monitor Istio service mesh health and configuration distribution."),
				mcp.WithString("namespace",
					mcp.Description("Namespace of the pod (optional). If specified, shows status for proxies in that namespace only."),
				),
				mcp.WithString("pod",
					mcp.Description("Pod name (optional, if not provided shows all proxies). Use this to check specific proxy sync status."),
				),
				mcp.WithTitleAnnotation("Istio: Proxy Status"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getProxyStatus,
		},
		{
			Tool: mcp.NewTool("get-istio-analyze",
				mcp.WithDescription("Analyze Istio configuration and report potential issues, misconfigurations, and best practice violations. This tool runs 'istioctl analyze' to provide comprehensive analysis of your Istio service mesh configuration."),
				mcp.WithString("namespace",
					mcp.Description("Namespace to analyze (optional). If specified, analyzes only the specified namespace. If not provided, analyzes the entire cluster."),
				),
				mcp.WithTitleAnnotation("Istio: Configuration Analysis"),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: s.getIstioAnalyze,
		},
	}
}

// Handler methods for networking tools
func (s *Server) getVirtualServices(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetVirtualServices(ctx, namespace)
	return NewTextResult(content, err), nil
}

func (s *Server) getDestinationRules(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetDestinationRules(ctx, namespace)
	return NewTextResult(content, err), nil
}

func (s *Server) getGateways(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetGateways(ctx, namespace)
	return NewTextResult(content, err), nil
}

func (s *Server) getServiceEntries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetServiceEntries(ctx, namespace)
	return NewTextResult(content, err), nil
}

// Handler methods for security tools
func (s *Server) getAuthorizationPolicies(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetAuthorizationPolicies(ctx, namespace)
	return NewTextResult(content, err), nil
}

func (s *Server) getPeerAuthentications(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetPeerAuthentications(ctx, namespace)
	return NewTextResult(content, err), nil
}

// Handler methods for configuration tools
func (s *Server) getEnvoyFilters(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetEnvoyFilters(ctx, namespace)
	return NewTextResult(content, err), nil
}

func (s *Server) getTelemetries(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetTelemetries(ctx, namespace)
	return NewTextResult(content, err), nil
}

func (s *Server) getIstioConfigSummary(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetIstioConfigSummary(ctx, namespace)
	return NewTextResult(content, err), nil
}

// Handler method for external dependency availability check
func (s *Server) checkExternalDependencyAvailability(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serviceName := ""
	if service := ctr.GetArguments()["service-name"]; service != nil {
		serviceName = service.(string)
	}
	if serviceName == "" {
		return NewTextResult("", fmt.Errorf("service-name is required")), nil
	}

	externalHost := ""
	if host := ctr.GetArguments()["external-host"]; host != nil {
		externalHost = host.(string)
	}
	if externalHost == "" {
		return NewTextResult("", fmt.Errorf("external-host is required")), nil
	}

	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}

	content, err := s.i.CheckExternalDependencyAvailability(ctx, serviceName, externalHost, namespace)
	return NewTextResult(content, err), nil
}

// Handler method for Istio namespace discovery
func (s *Server) discoverIstioNamespaces(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content, err := s.i.DiscoverNamespacesWithSidecars(ctx)
	return NewTextResult(content, err), nil
}

// Handler methods for proxy configuration tools
func (s *Server) getProxyClusters(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	podName := ""
	if pod := ctr.GetArguments()["pod"]; pod != nil {
		podName = pod.(string)
	}
	if podName == "" {
		return NewTextResult("", fmt.Errorf("pod name is required")), nil
	}
	content, err := s.i.ProxyConfig.GetClusters(ctx, namespace, podName)
	return NewTextResult(content, err), nil
}

func (s *Server) getProxyListeners(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	podName := ""
	if pod := ctr.GetArguments()["pod"]; pod != nil {
		podName = pod.(string)
	}
	if podName == "" {
		return NewTextResult("", fmt.Errorf("pod name is required")), nil
	}
	content, err := s.i.ProxyConfig.GetListeners(ctx, namespace, podName)
	return NewTextResult(content, err), nil
}

func (s *Server) getProxyRoutes(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	podName := ""
	if pod := ctr.GetArguments()["pod"]; pod != nil {
		podName = pod.(string)
	}
	if podName == "" {
		return NewTextResult("", fmt.Errorf("pod name is required")), nil
	}
	content, err := s.i.ProxyConfig.GetRoutes(ctx, namespace, podName)
	return NewTextResult(content, err), nil
}

func (s *Server) getProxyEndpoints(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	podName := ""
	if pod := ctr.GetArguments()["pod"]; pod != nil {
		podName = pod.(string)
	}
	if podName == "" {
		return NewTextResult("", fmt.Errorf("pod name is required")), nil
	}
	content, err := s.i.ProxyConfig.GetEndpoints(ctx, namespace, podName)
	return NewTextResult(content, err), nil
}

func (s *Server) getProxyBootstrap(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	podName := ""
	if pod := ctr.GetArguments()["pod"]; pod != nil {
		podName = pod.(string)
	}
	if podName == "" {
		return NewTextResult("", fmt.Errorf("pod name is required")), nil
	}
	content, err := s.i.ProxyConfig.GetBootstrap(ctx, namespace, podName)
	return NewTextResult(content, err), nil
}

func (s *Server) getProxyConfigDump(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	podName := ""
	if pod := ctr.GetArguments()["pod"]; pod != nil {
		podName = pod.(string)
	}
	if podName == "" {
		return NewTextResult("", fmt.Errorf("pod name is required")), nil
	}
	content, err := s.i.ProxyConfig.GetConfigDump(ctx, namespace, podName)
	return NewTextResult(content, err), nil
}

func (s *Server) getProxyStatus(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ""
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	podName := ""
	if pod := ctr.GetArguments()["pod"]; pod != nil {
		podName = pod.(string)
	}

	var content string
	var err error

	if podName != "" && namespace != "" {
		// Get status for specific pod
		content, err = s.i.ProxyConfig.GetProxyStatusForPod(ctx, namespace, podName)
	} else {
		// Get status for all proxies
		content, err = s.i.ProxyConfig.GetProxyStatus(ctx)
	}

	return NewTextResult(content, err), nil
}

// getIstioAnalyze performs Istio configuration analysis
func (s *Server) getIstioAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ""
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}

	content, err := s.i.ProxyConfig.GetAnalyze(ctx, namespace)
	return NewTextResult(content, err), nil
}

// Handler implementations (add to profile.go)
func (s *Server) getServices(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	content, err := s.i.GetServices(ctx, namespace)
	return NewTextResult(content, err), nil
}

func (s *Server) getPodsByService(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := "default"
	if ns := ctr.GetArguments()["namespace"]; ns != nil {
		namespace = ns.(string)
	}
	serviceName := ""
	if svc := ctr.GetArguments()["service"]; svc != nil {
		serviceName = svc.(string)
	}
	if serviceName == "" {
		return NewTextResult("", fmt.Errorf("service name is required - use 'get-services' first to discover available services")), nil
	}
	content, err := s.i.GetPodsByService(ctx, namespace, serviceName)
	return NewTextResult(content, err), nil
}

func init() {
	ProfileNames = make([]string, 0)
	for _, profile := range Profiles {
		ProfileNames = append(ProfileNames, profile.GetName())
	}
}
