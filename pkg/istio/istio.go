package istio

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
	"istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// AuthorizationHeader is the HTTP header name for authorization
const AuthorizationHeader = "Authorization"

// CloseWatchKubeConfig is a function type for closing kubeconfig watchers
type CloseWatchKubeConfig func() error

// Istio represents the Istio client and configuration
type Istio struct {
	kubeClient           kubernetes.Interface
	istioClient          versioned.Interface
	config               *rest.Config
	clientCmdConfig      clientcmd.ClientConfig
	CloseWatchKubeConfig CloseWatchKubeConfig
	ProxyConfig          *ProxyConfigClient
}

// NewIstio creates a new Istio client instance
func NewIstio(kubeconfig string) (*Istio, error) {
	config, clientCmdConfig, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	istioClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create istio client: %w", err)
	}

	return &Istio{
		kubeClient:      kubeClient,
		istioClient:     istioClient,
		config:          config,
		clientCmdConfig: clientCmdConfig,
		ProxyConfig:     NewProxyConfigClient(kubeconfig),
	}, nil
}

// buildConfig builds the Kubernetes configuration from kubeconfig path
func buildConfig(kubeconfig string) (*rest.Config, clientcmd.ClientConfig, error) {
	var clientCmdConfig clientcmd.ClientConfig
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, err
		}
		// Build clientCmdConfig for the specific kubeconfig file
		clientCmdConfig, err = clientcmd.NewClientConfigFromBytes(nil)
		if err != nil {
			// Fallback to loading from file path
			loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
			configOverrides := &clientcmd.ConfigOverrides{}
			clientCmdConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		}
	} else {
		// Try to use default kubeconfig locations (like kubectl does)
		// This will look for ~/.kube/config and other standard locations
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		clientCmdConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = clientCmdConfig.ClientConfig()
		if err != nil {
			return nil, nil, err
		}
	}

	return config, clientCmdConfig, nil
}

// GetVirtualServices retrieves Virtual Services from the specified namespace
func (i *Istio) GetVirtualServices(ctx context.Context, namespace string) (string, error) {
	vsList, err := i.istioClient.NetworkingV1alpha3().VirtualServices(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list virtual services: %w", err)
	}

	result := fmt.Sprintf("Found %d Virtual Services in namespace '%s':\n", len(vsList.Items), namespace)
	for _, vs := range vsList.Items {
		result += fmt.Sprintf("- %s\n", vs.Name)
		if vs.Spec.Hosts != nil {
			result += fmt.Sprintf("  Hosts: %v\n", vs.Spec.Hosts)
		}
		if vs.Spec.Gateways != nil {
			result += fmt.Sprintf("  Gateways: %v\n", vs.Spec.Gateways)
		}
		if len(vs.Spec.Http) > 0 {
			result += fmt.Sprintf("  HTTP Routes: %d\n", len(vs.Spec.Http))
		}
		if len(vs.Spec.Tcp) > 0 {
			result += fmt.Sprintf("  TCP Routes: %d\n", len(vs.Spec.Tcp))
		}
		if len(vs.Spec.Tls) > 0 {
			result += fmt.Sprintf("  TLS Routes: %d\n", len(vs.Spec.Tls))
		}
		result += "\n"
	}

	return result, nil
}

func (i *Istio) GetDestinationRules(ctx context.Context, namespace string) (string, error) {
	drList, err := i.istioClient.NetworkingV1alpha3().DestinationRules(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list destination rules: %w", err)
	}

	result := fmt.Sprintf("Found %d Destination Rules in namespace '%s':\n", len(drList.Items), namespace)
	for _, dr := range drList.Items {
		result += fmt.Sprintf("- %s\n", dr.Name)
		if dr.Spec.Host != "" {
			result += fmt.Sprintf("  Host: %s\n", dr.Spec.Host)
		}
	}
	return result, nil
}

func (i *Istio) GetGateways(ctx context.Context, namespace string) (string, error) {
	gwList, err := i.istioClient.NetworkingV1alpha3().Gateways(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list gateways: %w", err)
	}

	result := fmt.Sprintf("Found %d Gateways in namespace '%s':\n", len(gwList.Items), namespace)
	for _, gw := range gwList.Items {
		result += fmt.Sprintf("- %s\n", gw.Name)
		if gw.Spec.Selector != nil {
			result += fmt.Sprintf("  Selector: %v\n", gw.Spec.Selector)
		}
	}
	return result, nil
}

func (i *Istio) GetServiceEntries(ctx context.Context, namespace string) (string, error) {
	seList, err := i.istioClient.NetworkingV1alpha3().ServiceEntries(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list service entries: %w", err)
	}

	result := fmt.Sprintf("Found %d Service Entries in namespace '%s':\n", len(seList.Items), namespace)
	for _, se := range seList.Items {
		result += fmt.Sprintf("- %s\n", se.Name)
		if se.Spec.Hosts != nil {
			result += fmt.Sprintf("  Hosts: %v\n", se.Spec.Hosts)
		}
		if se.Spec.Location.String() != "" {
			result += fmt.Sprintf("  Location: %s\n", se.Spec.Location.String())
		}
	}
	return result, nil
}

// Security resources
func (i *Istio) GetAuthorizationPolicies(ctx context.Context, namespace string) (string, error) {
	apList, err := i.istioClient.SecurityV1beta1().AuthorizationPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list authorization policies: %w", err)
	}

	result := fmt.Sprintf("Found %d Authorization Policies in namespace '%s':\n", len(apList.Items), namespace)
	for _, ap := range apList.Items {
		result += fmt.Sprintf("- %s\n", ap.Name)
		if ap.Spec.Selector != nil && ap.Spec.Selector.MatchLabels != nil {
			result += fmt.Sprintf("  Selector: %v\n", ap.Spec.Selector.MatchLabels)
		}
		if ap.Spec.Action.String() != "" {
			result += fmt.Sprintf("  Action: %s\n", ap.Spec.Action.String())
		}
	}
	return result, nil
}

func (i *Istio) GetPeerAuthentications(ctx context.Context, namespace string) (string, error) {
	paList, err := i.istioClient.SecurityV1beta1().PeerAuthentications(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list peer authentications: %w", err)
	}

	result := fmt.Sprintf("Found %d Peer Authentications in namespace '%s':\n", len(paList.Items), namespace)
	for _, pa := range paList.Items {
		result += fmt.Sprintf("- %s\n", pa.Name)
		if pa.Spec.Selector != nil && pa.Spec.Selector.MatchLabels != nil {
			result += fmt.Sprintf("  Selector: %v\n", pa.Spec.Selector.MatchLabels)
		}
		if pa.Spec.Mtls != nil {
			result += fmt.Sprintf("  mTLS Mode: %s\n", pa.Spec.Mtls.Mode.String())
		}
	}
	return result, nil
}

// Configuration resources
func (i *Istio) GetEnvoyFilters(ctx context.Context, namespace string) (string, error) {
	efList, err := i.istioClient.NetworkingV1alpha3().EnvoyFilters(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list envoy filters: %w", err)
	}

	result := fmt.Sprintf("Found %d Envoy Filters in namespace '%s':\n", len(efList.Items), namespace)
	for _, ef := range efList.Items {
		result += fmt.Sprintf("- %s\n", ef.Name)
		if ef.Spec.WorkloadSelector != nil && ef.Spec.WorkloadSelector.Labels != nil {
			result += fmt.Sprintf("  Workload Selector: %v\n", ef.Spec.WorkloadSelector.Labels)
		}
	}
	return result, nil
}

func (i *Istio) GetTelemetries(ctx context.Context, namespace string) (string, error) {
	telList, err := i.istioClient.TelemetryV1alpha1().Telemetries(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list telemetries: %w", err)
	}

	result := fmt.Sprintf("Found %d Telemetry configurations in namespace '%s':\n", len(telList.Items), namespace)
	for _, tel := range telList.Items {
		result += fmt.Sprintf("- %s\n", tel.Name)
		if tel.Spec.Selector != nil && tel.Spec.Selector.MatchLabels != nil {
			result += fmt.Sprintf("  Selector: %v\n", tel.Spec.Selector.MatchLabels)
		}
	}
	return result, nil
}

func (i *Istio) GetIstioConfigSummary(ctx context.Context, namespace string) (string, error) {
	result := fmt.Sprintf("Istio Configuration Summary for namespace '%s':\n\n", namespace)

	// Get counts of each resource type
	vsList, err := i.istioClient.NetworkingV1alpha3().VirtualServices(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list virtual services: %v", err)
	} else {
		result += fmt.Sprintf("Virtual Services: %d\n", len(vsList.Items))
	}

	drList, err := i.istioClient.NetworkingV1alpha3().DestinationRules(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list destination rules: %v", err)
	} else {
		result += fmt.Sprintf("Destination Rules: %d\n", len(drList.Items))
	}

	gwList, err := i.istioClient.NetworkingV1alpha3().Gateways(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list gateways: %v", err)
	} else {
		result += fmt.Sprintf("Gateways: %d\n", len(gwList.Items))
	}

	seList, err := i.istioClient.NetworkingV1alpha3().ServiceEntries(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list service entries: %v", err)
	} else {
		result += fmt.Sprintf("Service Entries: %d\n", len(seList.Items))
	}

	apList, err := i.istioClient.SecurityV1beta1().AuthorizationPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list authorization policies: %v", err)
	} else {
		result += fmt.Sprintf("Authorization Policies: %d\n", len(apList.Items))
	}

	paList, err := i.istioClient.SecurityV1beta1().PeerAuthentications(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list peer authentications: %v", err)
	} else {
		result += fmt.Sprintf("Peer Authentications: %d\n", len(paList.Items))
	}

	efList, err := i.istioClient.NetworkingV1alpha3().EnvoyFilters(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list envoy filters: %v", err)
	} else {
		result += fmt.Sprintf("Envoy Filters: %d\n", len(efList.Items))
	}

	telList, err := i.istioClient.TelemetryV1alpha1().Telemetries(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to list telemetries: %v", err)
	} else {
		result += fmt.Sprintf("Telemetry Configurations: %d\n", len(telList.Items))
	}

	return result, nil
}

// CheckExternalDependencyAvailability checks if an external dependency is properly configured and accessible for a service
func (i *Istio) CheckExternalDependencyAvailability(ctx context.Context, serviceName, externalHost, namespace string) (string, error) {
	result := fmt.Sprintf("External Dependency Check for service '%s' -> '%s' in namespace '%s':\n\n", serviceName, externalHost, namespace)

	// Check 1: Service Entry existence
	seList, err := i.istioClient.NetworkingV1alpha3().ServiceEntries(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list service entries: %w", err)
	}

	serviceEntryFound := false
	var serviceEntryDetails string
	for _, se := range seList.Items {
		for _, host := range se.Spec.Hosts {
			if host == externalHost {
				serviceEntryFound = true
				serviceEntryDetails = fmt.Sprintf("[OK] Service Entry: '%s' found in namespace '%s'", se.Name, namespace)
				break
			}
		}
		if serviceEntryFound {
			break
		}
	}

	if !serviceEntryFound {
		// Check in istio-system namespace as well (common for global external dependencies)
		istioSystemSE, err := i.istioClient.NetworkingV1alpha3().ServiceEntries("istio-system").List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, se := range istioSystemSE.Items {
				for _, host := range se.Spec.Hosts {
					if host == externalHost {
						serviceEntryFound = true
						serviceEntryDetails = fmt.Sprintf("[OK] Service Entry: '%s' found in namespace 'istio-system' (global)", se.Name)
						break
					}
				}
				if serviceEntryFound {
					break
				}
			}
		}
	}

	if !serviceEntryFound {
		serviceEntryDetails = "[MISSING] Service Entry: Not found for external host"
	}

	// Check 2: Virtual Service routing
	vsList, err := i.istioClient.NetworkingV1alpha3().VirtualServices(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list virtual services: %w", err)
	}

	virtualServiceFound := false
	var virtualServiceDetails string
	for _, vs := range vsList.Items {
		for _, host := range vs.Spec.Hosts {
			if host == externalHost {
				virtualServiceFound = true
				virtualServiceDetails = fmt.Sprintf("[OK] Virtual Service: '%s' found with routing rules", vs.Name)
				break
			}
		}
		if virtualServiceFound {
			break
		}
	}

	if !virtualServiceFound {
		virtualServiceDetails = "[WARNING] Virtual Service: No specific routing rules found (may use default routing)"
	}

	// Check 3: Destination Rules
	drList, err := i.istioClient.NetworkingV1alpha3().DestinationRules(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list destination rules: %w", err)
	}

	destinationRuleFound := false
	var destinationRuleDetails string
	for _, dr := range drList.Items {
		if dr.Spec.Host == externalHost {
			destinationRuleFound = true
			destinationRuleDetails = fmt.Sprintf("[OK] Destination Rule: '%s' found with traffic policies", dr.Name)
			break
		}
	}

	if !destinationRuleFound {
		destinationRuleDetails = "[WARNING] Destination Rule: No specific traffic policies found (may use default policies)"
	}

	// Check 4: Authorization Policies
	apList, err := i.istioClient.SecurityV1beta1().AuthorizationPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list authorization policies: %w", err)
	}

	authorizationPolicyFound := false
	var authorizationPolicyDetails string
	for _, ap := range apList.Items {
		// Check if the policy allows access to the external host
		if ap.Spec.Action.String() == "ALLOW" {
			// This is a simplified check - in practice, you'd need to analyze the rules more carefully
			authorizationPolicyFound = true
			authorizationPolicyDetails = fmt.Sprintf("[OK] Authorization Policy: '%s' found (ALLOW action)", ap.Name)
			break
		}
	}

	if !authorizationPolicyFound {
		authorizationPolicyDetails = "[WARNING] Authorization Policy: No explicit ALLOW policies found (may use default allow)"
	}

	// Build the result
	result += serviceEntryDetails + "\n"
	result += virtualServiceDetails + "\n"
	result += destinationRuleDetails + "\n"
	result += authorizationPolicyDetails + "\n\n"

	// Overall assessment
	if serviceEntryFound {
		result += "[RESULT] External dependency '" + externalHost + "' is available for service '" + serviceName + "'\n"
		result += "   The Service Entry exists, which means the external service is registered in the mesh.\n"
		if virtualServiceFound || destinationRuleFound {
			result += "   Additional routing and traffic policies are configured.\n"
		}
	} else {
		result += "[RESULT] External dependency '" + externalHost + "' is NOT available for service '" + serviceName + "'\n"
		result += "   You need to create a Service Entry for '" + externalHost + "' before the service can access it.\n"
		result += "   Consider creating it in namespace '" + namespace + "' or globally in 'istio-system'.\n"
	}

	return result, nil
}

// GetServices retrieves all Kubernetes services in a namespace
func (i *Istio) GetServices(ctx context.Context, namespace string) (string, error) {
	services, err := i.kubeClient.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list services: %w", err)
	}

	result := fmt.Sprintf("Services in namespace '%s':\n\n", namespace)
	result += fmt.Sprintf("Found %d services:\n\n", len(services.Items))

	if len(services.Items) == 0 {
		result += "No services found in this namespace.\n"
		return result, nil
	}

	// Group services by type for better organization
	var clusterIPServices []string
	var nodePortServices []string
	var loadBalancerServices []string
	var headlessServices []string

	for _, service := range services.Items {
		serviceLine := fmt.Sprintf("%-30s", service.Name)

		// Add service type and cluster IP info
		switch service.Spec.Type {
		case "NodePort":
			nodePortServices = append(nodePortServices, fmt.Sprintf("%s (NodePort: %s)", serviceLine, service.Spec.ClusterIP))
		case "LoadBalancer":
			externalIP := "<pending>"
			if len(service.Status.LoadBalancer.Ingress) > 0 {
				if service.Status.LoadBalancer.Ingress[0].IP != "" {
					externalIP = service.Status.LoadBalancer.Ingress[0].IP
				} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
					externalIP = service.Status.LoadBalancer.Ingress[0].Hostname
				}
			}
			loadBalancerServices = append(loadBalancerServices, fmt.Sprintf("%s (LoadBalancer: %s)", serviceLine, externalIP))
		default:
			if service.Spec.ClusterIP == "None" {
				headlessServices = append(headlessServices, fmt.Sprintf("%s (Headless)", serviceLine))
			} else {
				clusterIPServices = append(clusterIPServices, fmt.Sprintf("%s (ClusterIP: %s)", serviceLine, service.Spec.ClusterIP))
			}
		}
	}

	// Output organized by service type
	if len(clusterIPServices) > 0 {
		result += " ClusterIP Services:\n"
		for _, svc := range clusterIPServices {
			result += fmt.Sprintf("   %s\n", svc)
		}
		result += "\n"
	}

	if len(nodePortServices) > 0 {
		result += " NodePort Services:\n"
		for _, svc := range nodePortServices {
			result += fmt.Sprintf("   %s\n", svc)
		}
		result += "\n"
	}

	if len(loadBalancerServices) > 0 {
		result += " LoadBalancer Services:\n"
		for _, svc := range loadBalancerServices {
			result += fmt.Sprintf("   %s\n", svc)
		}
		result += "\n"
	}

	if len(headlessServices) > 0 {
		result += " Headless Services:\n"
		for _, svc := range headlessServices {
			result += fmt.Sprintf("   %s\n", svc)
		}
		result += "\n"
	}

	result += "Next step: Use 'get-pods-by-service' to find pods backing any of these services\n"
	result += "   Example: get-pods-by-service --namespace " + namespace + " --service <service-name>\n"

	return result, nil
}

// GetPodsByService finds pods backing a specific Kubernetes service
func (i *Istio) GetPodsByService(ctx context.Context, namespace, serviceName string) (string, error) {
	// Get the service to find its selector
	service, err := i.kubeClient.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	result := fmt.Sprintf("Pods backing service '%s' in namespace '%s':\n\n", serviceName, namespace)

	// Handle headless services or services without selectors
	if service.Spec.Selector == nil {
		result += fmt.Sprintf("  Service '%s' has no selector - this is likely:\n", serviceName)
		result += "   - A headless service with manual endpoints\n"
		result += "   - An external service (ExternalName type)\n"
		result += "   - A service with manually configured endpoints\n\n"

		// Try to get endpoints to show what's configured
		endpoints, err := i.kubeClient.CoreV1().Endpoints(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err == nil && len(endpoints.Subsets) > 0 {
			result += " Configured endpoints:\n"
			for _, subset := range endpoints.Subsets {
				for _, addr := range subset.Addresses {
					if addr.TargetRef != nil && addr.TargetRef.Kind == "Pod" {
						result += fmt.Sprintf("   - Pod: %s (IP: %s)\n", addr.TargetRef.Name, addr.IP)
					} else {
						result += fmt.Sprintf("   - IP: %s\n", addr.IP)
					}
				}
			}
		}
		return result, nil
	}

	// Convert selector to label selector string
	var selectorParts []string
	for key, value := range service.Spec.Selector {
		selectorParts = append(selectorParts, fmt.Sprintf("%s=%s", key, value))
	}
	labelSelector := strings.Join(selectorParts, ",")

	// Find pods matching the service selector
	pods, err := i.kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods for service %s: %w", serviceName, err)
	}

	// Separate running and non-running pods
	var runningPods []v1.Pod
	var nonRunningPods []v1.Pod

	for _, pod := range pods.Items {
		if pod.Status.Phase == "Running" {
			runningPods = append(runningPods, pod)
		} else {
			nonRunningPods = append(nonRunningPods, pod)
		}
	}

	result += fmt.Sprintf(" Service selector: %s\n", labelSelector)
	result += fmt.Sprintf(" Total pods found: %d (%d running, %d not running)\n\n",
		len(pods.Items), len(runningPods), len(nonRunningPods))

	// Show running pods (most important)
	if len(runningPods) > 0 {
		result += fmt.Sprintf(" Running pods (%d) - Ready for proxy commands:\n", len(runningPods))
		for _, pod := range runningPods {
			// Check if it has Istio sidecar
			hasIstio := false
			for _, container := range pod.Spec.Containers {
				if container.Name == "istio-proxy" {
					hasIstio = true
					break
				}
			}

			readyIcon := "❌"
			if isPodReady(pod) {
				readyIcon = "✅"
			}

			istioIcon := "🔗"
			if hasIstio {
				istioIcon = "🕸️"
			}

			result += fmt.Sprintf("   %s %s %s\n", readyIcon, istioIcon, pod.Name)
			result += fmt.Sprintf("      IP: %-15s Node: %s\n", pod.Status.PodIP, pod.Spec.NodeName)

			// Show main application containers (exclude istio-proxy)
			var appContainers []string
			for _, container := range pod.Spec.Containers {
				if container.Name != "istio-proxy" {
					appContainers = append(appContainers, container.Name)
				}
			}
			result += fmt.Sprintf("      Containers: %s\n", strings.Join(appContainers, ", "))

			if hasIstio {
				result += fmt.Sprintf("      🕸️  Istio mesh: ENABLED\n")
			} else {
				result += fmt.Sprintf("      ⚠️  Istio mesh: NOT ENABLED\n")
			}
			result += "\n"
		}
	}

	// Show non-running pods for completeness
	if len(nonRunningPods) > 0 {
		result += fmt.Sprintf("⏳ Non-running pods (%d):\n", len(nonRunningPods))
		for _, pod := range nonRunningPods {
			result += fmt.Sprintf("   ❌ %s (Status: %s)\n", pod.Name, pod.Status.Phase)
		}
		result += "\n"
	}

	if len(runningPods) == 0 {
		result += "⚠️  No running pods found backing this service!\n"
		result += "💡 This could mean:\n"
		result += "   - The deployment is scaled to 0 replicas\n"
		result += "   - Pods are failing to start\n"
		result += "   - Label selector mismatch between service and pods\n\n"
		return result, nil
	}

	// Add helpful next steps
	result += "💡 Next steps - Use these pod names with proxy commands:\n"
	if len(runningPods) > 0 {
		examplePod := runningPods[0].Name
		result += fmt.Sprintf("   get-proxy-status --namespace %s --pod %s\n", namespace, examplePod)
		result += fmt.Sprintf("   get-proxy-clusters --namespace %s --pod %s\n", namespace, examplePod)
		result += fmt.Sprintf("   get-proxy-listeners --namespace %s --pod %s\n", namespace, examplePod)
		result += fmt.Sprintf("   get-proxy-routes --namespace %s --pod %s\n", namespace, examplePod)
	}

	return result, nil
}

// Helper function to check if pod is ready (already exists but ensuring it's here)
func isPodReady(pod v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}

// DiscoverNamespacesWithSidecars finds namespaces that have pods with Istio sidecars
// and returns them sorted by the number of sidecars (most injected first)
func (i *Istio) DiscoverNamespacesWithSidecars(ctx context.Context) (string, error) {
	namespacesWithSidecars := make(map[string]int)

	// Get running pods only (server-side filtering)
	pods, err := i.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return "", fmt.Errorf("failed to list running pods for Istio sidecar discovery: %w", err)
	}

	// Count sidecars per namespace
	for _, pod := range pods.Items {
		// Skip pods that are not running or have no containers
		if pod.Status.Phase != "Running" || len(pod.Spec.Containers) == 0 {
			continue
		}

		// Check if pod has istio-proxy sidecar
		for _, container := range pod.Spec.Containers {
			if container.Name == "istio-proxy" {
				namespacesWithSidecars[pod.Namespace]++
				break
			}
		}
	}

	if len(namespacesWithSidecars) == 0 {
		return "No namespaces with Istio sidecars found", nil
	}

	// Create a slice of namespace counts for sorting
	type namespaceCount struct {
		namespace string
		count     int
	}

	var namespaceCounts []namespaceCount
	for ns, count := range namespacesWithSidecars {
		namespaceCounts = append(namespaceCounts, namespaceCount{namespace: ns, count: count})
	}

	// Sort by count (descending) and then by namespace name (ascending)
	sort.Slice(namespaceCounts, func(i, j int) bool {
		if namespaceCounts[i].count != namespaceCounts[j].count {
			return namespaceCounts[i].count > namespaceCounts[j].count
		}
		return namespaceCounts[i].namespace < namespaceCounts[j].namespace
	})

	// Build result string
	result := fmt.Sprintf("Found %d namespaces with Istio sidecars:\n\n", len(namespaceCounts))
	result += "Rank | Namespace | Sidecar Count | Recommendation\n"
	result += "-----|-----------|---------------|----------------\n"

	for rank, nc := range namespaceCounts {
		var recommendation string
		if rank == 0 {
			recommendation = "BEST - Most Istio-injected workloads"
		} else if rank < 3 { // 3 is arbitrary, adjust as needed
			recommendation = "Good - High Istio adoption"
		} else if rank < 5 {
			recommendation = "Moderate - Some Istio usage"
		} else {
			recommendation = "Low - Minimal Istio usage"
		}

		result += fmt.Sprintf("%4d | %-9s | %13d | %s\n", rank+1, nc.namespace, nc.count, recommendation)
	}

	result += "\n💡 **Recommendation**: Start with the top-ranked namespace for Istio operations as it likely contains the most Istio configuration and traffic."

	return result, nil
}

func (i *Istio) WatchKubeConfig(onKubeConfigChange func() error) {
	if i.clientCmdConfig == nil {
		klog.V(1).Info("No client config available for kubeconfig watching")
		return
	}
	configAccess := i.clientCmdConfig.ConfigAccess()
	if configAccess == nil {
		klog.V(1).Info("No config access available for kubeconfig watching")
		return
	}
	kubeConfigFiles := configAccess.GetLoadingPrecedence()
	if len(kubeConfigFiles) == 0 {
		klog.V(1).Info("No kubeconfig files found for watching")
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Errorf("Failed to create kubeconfig watcher: %v", err)
		return
	}
	for _, file := range kubeConfigFiles {
		err := watcher.Add(file)
		if err != nil {
			klog.Warningf("Failed to watch kubeconfig file %s: %v", file, err)
		} else {
			klog.V(2).Infof("Watching kubeconfig file: %s", file)
		}
	}
	go func() {
		defer func() {
			if err := watcher.Close(); err != nil {
				klog.Errorf("Failed to close kubeconfig watcher: %v", err)
			}
		}()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					klog.V(1).Info("Kubeconfig watcher events channel closed")
					return
				}
				klog.V(2).Infof("Kubeconfig file changed: %s (event: %s)", event.Name, event.Op)
				if err := onKubeConfigChange(); err != nil {
					klog.Errorf("Failed to handle kubeconfig change: %v", err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					klog.V(1).Info("Kubeconfig watcher errors channel closed")
					return
				}
				klog.Errorf("Kubeconfig watcher error: %v", err)
			}
		}
	}()
	if i.CloseWatchKubeConfig != nil {
		if err := i.CloseWatchKubeConfig(); err != nil {
			klog.Errorf("Failed to close previous kubeconfig watcher: %v", err)
		}
	}
	i.CloseWatchKubeConfig = watcher.Close
}

func (i *Istio) Close() {
	if i.CloseWatchKubeConfig != nil {
		if err := i.CloseWatchKubeConfig(); err != nil {
			klog.Errorf("Failed to close kubeconfig watcher: %v", err)
		}
	}
	klog.V(1).Info("Closing Istio client connections")
}
