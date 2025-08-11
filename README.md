# Istio MCP Server

A **Model Context Protocol (MCP) server** that provides AI assistants and developers with **read-only access** to Istio service mesh resources in Kubernetes clusters. This server enables intelligent querying of Istio configurations, Virtual Services, Destination Rules, Gateways, and Envoy proxy configurations through a safe, non-destructive interface.

## 🚀 Overview

The Istio MCP Server bridges the gap between AI assistants and Istio service mesh operations by implementing the Model Context Protocol. It provides comprehensive tools for querying Istio resources including Virtual Services, Destination Rules, Gateways, and proxy configurations **without any risk of modifying or deleting resources**.

**Key Benefits:**
- 🔒 **100% Read-Only Operations** - No destructive commands allowed
- 🤖 **AI Assistant Friendly** - Designed for MCP protocol integration
- 🔍 **Comprehensive Istio Access** - Covers all major Istio resource types
- 🛡️ **Safe by Design** - Zero risk of accidental resource modifications
- 🌐 **Multi-Protocol Support** - STDIO, SSE, and HTTP protocols
- 📊 **Rich Observability** - Access to Envoy proxy configurations and telemetry

## ✨ Features

### 🔧 Core Istio Resources (Read-Only)
- **Virtual Services**: Query Istio Virtual Services and routing rules
- **Destination Rules**: Query Istio Destination Rules and traffic policies
- **Gateways**: Query Istio Gateways and ingress configurations
- **Service Entries**: Query Istio Service Entries and external services
- **Envoy Filters**: Query Istio Envoy Filters and custom configurations

### 🛡️ Security & Policies (Read-Only)
- **Authorization Policies**: Query Istio Authorization Policies
- **Peer Authentications**: Query Istio Peer Authentication policies
- **Security Configurations**: Access Istio security settings

### 📊 Observability & Telemetry (Read-Only)
- **Telemetry Configurations**: Query Istio telemetry settings
- **Proxy Status**: Get Envoy proxy health and status information
- **Configuration Summaries**: Comprehensive Istio configuration overviews

### 🌐 Envoy Proxy Access (Read-Only)
- **Cluster Configuration**: Access Envoy cluster configurations
- **Listener Configuration**: Access Envoy listener configurations
- **Route Configuration**: Access Envoy route configurations
- **Endpoint Configuration**: Access Envoy endpoint configurations
- **Bootstrap Configuration**: Access Envoy bootstrap configurations
- **Full Configuration Dumps**: Complete Envoy configuration snapshots

## 🚀 Getting Started

### Prerequisites

- **Go 1.24+** for building from source
- **Kubernetes cluster** with Istio installed
- **kubectl** configured with appropriate permissions
- **istioctl** installed (for advanced proxy configuration features)

### Installation

```bash
# Install via npm (recommended)
npm install -g istio-mcp-server

# Or build from source
git clone https://github.com/krutsko/istio-mcp-server.git
cd istio-mcp-server
make build
```

### Claude Desktop

#### Using npx

If you have npm installed, this is the fastest way to get started with `istio-mcp-server` on Claude Desktop.

Open your `claude_desktop_config.json` and add the mcp server to the list of `mcpServers`:
``` json
{
  "mcpServers": {
    "istio": {
      "command": "npx",
      "args": [
        "-y",
        "istio-mcp-server@latest"
      ]
    }
  }
}
```

### VS Code / VS Code Insiders

Install the Istio MCP server extension in VS Code Insiders by pressing the following link:

[<img src="https://img.shields.io/badge/VS_Code-VS_Code?style=flat-square&label=Install%20Server&color=0098FF" alt="Install in VS Code">](https://insiders.vscode.dev/redirect?url=vscode%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522istio%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522istio-mcp-server%2540latest%2522%255D%257D)
[<img alt="Install in VS Code Insiders" src="https://img.shields.io/badge/VS_Code_Insiders-VS_Code_Insiders?style=flat-square&label=Install%20Server&color=24bfa5">](https://insiders.vscode.dev/redirect?url=vscode-insiders%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522istio%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522istio-mcp-server%2540latest%2522%255D%257D)

Alternatively, you can install the extension manually by running the following command:

```shell
# For VS Code
code --add-mcp '{"name":"istio","command":"npx","args":["istio-mcp-server@latest"]}'
# For VS Code Insiders
code-insiders --add-mcp '{"name":"istio","command":"npx","args":["istio-mcp-server@latest"]}'
```

### Cursor

Install the Istio MCP server extension in Cursor by pressing the following link:

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](cursor://anysphere.cursor-deeplink/mcp/install?name=istio-mcp-server&config=eyJjb21tYW5kIjoibnB4IC15IGlzdGlvLW1jcC1zZXJ2ZXJAbGF0ZXN0In0%3D)


Alternatively, you can install the extension manually by editing the `mcp.json` file:

```json
{
  "mcpServers": {
    "istio-mcp-server": {
      "command": "npx",
      "args": ["-y", "istio-mcp-server@latest"]
    }
  }
}
```

### Goose CLI

[Goose CLI](https://blog.marcnuri.com/goose-on-machine-ai-agent-cli-introduction) is the easiest (and cheapest) way to get rolling with artificial intelligence (AI) agents.

#### Using npm

If you have npm installed, this is the fastest way to get started with `istio-mcp-server`.

Open your goose `config.yaml` and add the mcp server to the list of `mcpServers`:
```yaml
extensions:
  istio:
    command: npx
    args:
      - -y
      - istio-mcp-server@latest
```

### Basic Usage

```bash
# Run in STDIO mode (for MCP clients)
./bin/istio-mcp-server --kubeconfig ~/.kube/config

# Run SSE server on port 8080
./bin/istio-mcp-server --sse-port 8080

# Run HTTP server on port 8080
./bin/istio-mcp-server --http-port 8080

# Show all available options
./bin/istio-mcp-server --help
```

## 🛠️ Available Tools

### 🌐 Networking Resources
- `get-virtual-services` - List Virtual Services in a namespace
- `get-destination-rules` - List Destination Rules in a namespace  
- `get-gateways` - List Gateways in a namespace
- `get-service-entries` - List Service Entries in a namespace

### 🛡️ Security Resources
- `get-authorization-policies` - List Authorization Policies in a namespace
- `get-peer-authentications` - List Peer Authentications in a namespace

### ⚙️ Configuration Resources
- `get-envoy-filters` - List Envoy Filters in a namespace
- `get-telemetry` - List Telemetry configurations in a namespace
- `get-istio-config` - Get comprehensive Istio configuration summary

### 🔍 Proxy Configuration
- `get-proxy-clusters` - Get Envoy cluster configuration from a pod
- `get-proxy-listeners` - Get Envoy listener configuration from a pod
- `get-proxy-routes` - Get Envoy route configuration from a pod
- `get-proxy-endpoints` - Get Envoy endpoint configuration from a pod
- `get-proxy-bootstrap` - Get Envoy bootstrap configuration from a pod
- `get-proxy-config-dump` - Get full Envoy configuration dump from a pod
- `get-proxy-status` - Get proxy status information

## ⚙️ Configuration

The server supports various configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `--kubeconfig` | Path to kubeconfig file | `~/.kube/config` |
| `--sse-port` | Start SSE server on specified port | Disabled |
| `--http-port` | Start HTTP server on specified port | Disabled |
| `--log-level` | Set logging level (0-9) | `0` |
| `--profile` | MCP profile to use | `"full"` |

**🔒 Security Note**: This server operates in read-only mode by design. All operations are safe and non-destructive.

## 🏗️ Architecture

The Istio MCP Server follows clean architecture principles with clear separation of concerns:

```
istio-mcp-server/
├── cmd/                    # Application entrypoints and CLI commands
├── pkg/
│   ├── istio-mcp-server/  # Core application logic and CLI handling
│   ├── istio/             # Istio client and resource management
│   ├── mcp/               # MCP server implementation and tool definitions
│   ├── version/           # Version information and build metadata
│   └── output/            # Output formatting and display utilities
└── npm/                   # NPM package distribution
```

## 🧪 Development

```bash
# Install development dependencies
make deps

# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Clean build artifacts
make clean

# Build for all platforms
make build-all-platforms

# Test release process
make test-release
```

## 🔗 Related Projects

- **[Model Context Protocol](https://modelcontextprotocol.io/)** - The protocol specification
- **[Istio](https://istio.io/)** - Service mesh platform
- **[Kubernetes](https://kubernetes.io/)** - Container orchestration platform
- **[Envoy Proxy](https://www.envoyproxy.io/)** - High-performance proxy

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE.md) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.


