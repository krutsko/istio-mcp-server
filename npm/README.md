# Istio MCP Server NPM Packages

This directory contains the npm packages for the Istio MCP Server.

## Package Structure

- `istio-mcp-server/` - Main package with platform detection
- `istio-mcp-server-darwin-amd64/` - macOS Intel binary
- `istio-mcp-server-darwin-arm64/` - macOS Apple Silicon binary
- `istio-mcp-server-linux-amd64/` - Linux Intel binary
- `istio-mcp-server-linux-arm64/` - Linux ARM64 binary
- `istio-mcp-server-windows-amd64/` - Windows Intel binary
- `istio-mcp-server-windows-arm64/` - Windows ARM64 binary

## Installation

```bash
npm install istio-mcp-server
```

The main package will automatically install the appropriate platform-specific binary.

## Usage

```bash
npx istio-mcp-server
```

## Development

To build and publish these packages:

```bash
make npm-publish
```

This will:
1. Build binaries for all platforms
2. Copy binaries to respective npm packages
3. Update versions in package.json files
4. Publish all packages to npm
