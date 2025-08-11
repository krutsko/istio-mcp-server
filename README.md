# Istio MCP Server

A Model Context Protocol (MCP) server for Istio service mesh operations.

## Overview

This project provides an MCP server that allows AI assistants and other MCP clients to interact with Istio service mesh resources in Kubernetes clusters. **This server is designed for read-only operations only - no destructive commands are allowed.** It provides comprehensive tools for querying Istio resources including Virtual Services, Destination Rules, Gateways, and proxy configurations without any risk of modifying or deleting resources.

## Features

- **Virtual Services**: Query Istio Virtual Services (read-only)
- **Destination Rules**: Query Istio Destination Rules (read-only)
- **Gateways**: Query Istio Gateways (read-only)
- **Service Entries**: Query Istio Service Entries (read-only)
- **Security Policies**: Query Authorization Policies and Peer Authentications (read-only)
- **Proxy Configuration**: Access Envoy proxy configurations including clusters, listeners, routes, and endpoints (read-only)
- **Configuration Management**: Get comprehensive Istio configuration summaries (read-only)
- **Multiple Protocols**: Support for STDIO, SSE, and HTTP protocols
- **Safe Operations**: All operations are read-only with no destructive commands allowed

## Quick Start

### Prerequisites

- Go 1.24 or later
- Access to a Kubernetes cluster with Istio installed
- kubectl configured with appropriate permissions
- istioctl installed (for proxy configuration features)

### Build and Run

```bash
# Initialize dependencies
make deps

# Build the server
make build

# Run the server (STDIO mode)
make run

# Or run with specific options
./bin/istio-mcp-server --kubeconfig ~/.kube/config

# Run SSE server on port 8080
./bin/istio-mcp-server --sse-port 8080

# Show help
./bin/istio-mcp-server --help
```

### Available Tools

#### Networking Resources
- `get-virtual-services` - List Virtual Services in a namespace
- `get-destination-rules` - List Destination Rules in a namespace  
- `get-gateways` - List Gateways in a namespace
- `get-service-entries` - List Service Entries in a namespace

#### Security Resources
- `get-authorization-policies` - List Authorization Policies in a namespace
- `get-peer-authentications` - List Peer Authentications in a namespace

#### Configuration Resources
- `get-envoy-filters` - List Envoy Filters in a namespace
- `get-telemetry` - List Telemetry configurations in a namespace
- `get-istio-config` - Get comprehensive Istio configuration summary

#### Proxy Configuration
- `get-proxy-clusters` - Get Envoy cluster configuration from a pod
- `get-proxy-listeners` - Get Envoy listener configuration from a pod
- `get-proxy-routes` - Get Envoy route configuration from a pod
- `get-proxy-endpoints` - Get Envoy endpoint configuration from a pod
- `get-proxy-bootstrap` - Get Envoy bootstrap configuration from a pod
- `get-proxy-config-dump` - Get full Envoy configuration dump from a pod
- `get-proxy-status` - Get proxy status information

## Configuration

The server supports various configuration options:

- `--kubeconfig`: Path to kubeconfig file
- `--sse-port`: Start SSE server on specified port
- `--http-port`: Start HTTP server on specified port
- `--log-level`: Set logging level (0-9)
- `--profile`: MCP profile to use (default: "full")

**Note**: This server operates in read-only mode by design. All operations are safe and non-destructive.

## Development

```bash
# Install development dependencies
make dev-deps

# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Clean build artifacts
make clean
```

## Architecture

The Istio MCP Server is built with a clean, modular architecture:

- **cmd/**: Application entrypoints and CLI commands
- **pkg/istio-mcp-server/**: Core application logic and CLI handling
- **pkg/istio/**: Istio client and resource management
- **pkg/mcp/**: MCP server implementation and tool definitions
- **pkg/version/**: Version information
- **pkg/output/**: Output formatting utilities

The server follows Go best practices with proper error handling, context propagation, and clean separation of concerns.

## License

This project is licensed under the MIT License.

