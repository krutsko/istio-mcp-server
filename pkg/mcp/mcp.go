package mcp

import (
	"context"
	"net/http"

	"github.com/krutsko/istio-mcp-server/pkg/istio"
	"github.com/krutsko/istio-mcp-server/pkg/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Configuration holds the server configuration
type Configuration struct {
	Profile    Profile
	Kubeconfig string
}

// Server represents the Istio MCP server
type Server struct {
	configuration *Configuration
	server        *server.MCPServer
	i             *istio.Istio
}

// NewServer creates a new Istio MCP server instance
func NewServer(configuration Configuration) (*Server, error) {
	s := &Server{
		configuration: &configuration,
		server: server.NewMCPServer(
			version.BinaryName,
			version.Version,
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
		),
	}
	if err := s.reloadIstioClient(); err != nil {
		return nil, err
	}
	s.i.WatchKubeConfig(s.reloadIstioClient)
	return s, nil
}

// reloadIstioClient reloads the Istio client and updates the server tools
func (s *Server) reloadIstioClient() error {
	i, err := istio.NewIstio(s.configuration.Kubeconfig)
	if err != nil {
		return err
	}
	s.i = i
	// All tools are read-only and non-destructive, so no filtering needed
	tools := s.configuration.Profile.GetTools(s)
	s.server.SetTools(tools...)
	return nil
}

// ServeStdio starts the server in STDIO mode
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}

// ServeSse creates and returns an SSE server instance
func (s *Server) ServeSse(baseUrl string) *server.SSEServer {
	options := make([]server.SSEOption, 0)
	options = append(options, server.WithSSEContextFunc(contextFunc))
	if baseUrl != "" {
		options = append(options, server.WithBaseURL(baseUrl))
	}
	return server.NewSSEServer(s.server, options...)
}

// ServeHTTP creates and returns a streaming HTTP server instance
func (s *Server) ServeHTTP() *server.StreamableHTTPServer {
	options := []server.StreamableHTTPOption{
		server.WithHTTPContextFunc(contextFunc),
	}
	return server.NewStreamableHTTPServer(s.server, options...)
}

// Close cleans up server resources
func (s *Server) Close() {
	if s.i != nil {
		s.i.Close()
	}
}

// NewTextResult creates a new text result for tool responses
func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: err.Error(),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}

// contextFunc adds authorization header to context for HTTP requests
func contextFunc(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, istio.AuthorizationHeader, r.Header.Get(istio.AuthorizationHeader))
}
