package istio

import (
	"context"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"istio.io/client-go/pkg/clientset/versioned"
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
