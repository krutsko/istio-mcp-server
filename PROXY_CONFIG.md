# Istio Proxy Configuration Support

This document describes the proxy configuration features available in the Istio MCP Server.

## Overview

The Istio MCP Server provides comprehensive tools for accessing Envoy proxy configurations within Istio service mesh. These tools allow you to inspect the runtime configuration of Envoy proxies running in your Istio-managed pods.

## Available Tools

The Istio MCP Server supports these proxy configuration tools:

- **get-proxy-clusters**: Get Envoy cluster configuration from a pod
- **get-proxy-listeners**: Get Envoy listener configuration from a pod  
- **get-proxy-routes**: Get Envoy route configuration from a pod
- **get-proxy-endpoints**: Get Envoy endpoint configuration from a pod
- **get-proxy-bootstrap**: Get Envoy bootstrap configuration from a pod
- **get-proxy-config-dump**: Get full Envoy configuration dump from a pod
- **get-proxy-status**: Get proxy status information for all pods or a specific pod

## Implementation Details

- Uses `istioctl proxy-config` commands under the hood
- Requires `istioctl` to be installed on the system
- Returns JSON formatted output for easy parsing
- Includes proper error handling and timeouts
- Supports both namespace-wide and pod-specific queries

## Usage

Each tool requires:
- `namespace` (optional, defaults to 'default')
- `pod` (required for most tools, except `get-proxy-status`)

### Examples

```bash
# Get cluster configuration for a specific pod
get-proxy-clusters --namespace default --pod my-app-pod

# Get listener configuration
get-proxy-listeners --namespace istio-system --pod istio-ingressgateway-xyz

# Get route configuration
get-proxy-routes --namespace default --pod frontend-service

# Get proxy status for all pods in a namespace
get-proxy-status --namespace default

# Get proxy status for a specific pod
get-proxy-status --namespace default --pod my-app-pod
```

## Prerequisites

- Istio installed in your Kubernetes cluster
- `istioctl` CLI tool installed and configured
- Proper RBAC permissions to access pod information
- Network access to the Kubernetes API server

## Error Handling

The tools include comprehensive error handling for common scenarios:
- Pod not found
- Istio proxy not running in pod
- Network connectivity issues
- Permission denied errors
- Invalid namespace or pod names
