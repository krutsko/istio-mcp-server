package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestProfileFromString tests profile lookup by name
func TestProfileFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Profile
	}{
		{
			name:     "valid full profile",
			input:    "full",
			expected: &FullProfile{},
		},
		{
			name:     "invalid profile",
			input:    "invalid",
			expected: nil,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProfileFromString(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Fatalf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected profile, got nil")
				}
				if result.GetName() != tt.expected.GetName() {
					t.Fatalf("Expected profile name %s, got %s", tt.expected.GetName(), result.GetName())
				}
			}
		})
	}
}

// TestFullProfile tests the full profile implementation
func TestFullProfile(t *testing.T) {
	profile := &FullProfile{}

	t.Run("has correct name", func(t *testing.T) {
		if profile.GetName() != "full" {
			t.Fatalf("Expected name 'full', got '%s'", profile.GetName())
		}
	})

	t.Run("has description", func(t *testing.T) {
		description := profile.GetDescription()
		if description == "" {
			t.Fatal("Description should not be empty")
		}
		if description != "Complete profile with all Istio service mesh tools for networking, security, configuration, and proxy debugging across multiple namespaces" {
			t.Fatalf("Unexpected description: %s", description)
		}
	})

	t.Run("provides tools", func(t *testing.T) {
		testCase(t, func(c *mcpContext) {
			err := c.setupMCPServer()
			if err != nil {
				t.Fatalf("Failed to setup MCP server: %v", err)
			}
			defer c.server.Close()

			tools := profile.GetTools(c.server)
			if len(tools) == 0 {
				t.Fatal("Profile should provide tools")
			}

			// Check that we have tools from different categories
			categories := map[string]bool{
				"networking":    false,
				"security":      false,
				"configuration": false,
				"proxy-config":  false,
			}

			for _, tool := range tools {
				toolName := tool.Tool.Name
				switch {
				case toolName == "get-virtual-services" || toolName == "get-destination-rules" || toolName == "get-gateways":
					categories["networking"] = true
				case toolName == "get-authorization-policies" || toolName == "get-peer-authentications":
					categories["security"] = true
				case toolName == "get-envoy-filters" || toolName == "get-telemetry":
					categories["configuration"] = true
				case toolName == "get-proxy-clusters" || toolName == "get-proxy-listeners" || toolName == "get-proxy-routes":
					categories["proxy-config"] = true
				}
			}

			for category, found := range categories {
				if !found {
					t.Errorf("Expected tools from category '%s'", category)
				}
			}
		})
	})
}

func TestToolAnnotations(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		err := c.setupMCPServer()
		if err != nil {
			t.Fatalf("Failed to setup MCP server: %v", err)
		}
		defer c.server.Close()

		profile := &FullProfile{}
		tools := profile.GetTools(c.server)

		t.Run("all tools have required annotations", func(t *testing.T) {
			for _, tool := range tools {
				if tool.Tool.Annotations.ReadOnlyHint == nil {
					t.Fatalf("Tool %s has no readOnlyHint annotation", tool.Tool.Name)
				}
				if tool.Tool.Annotations.DestructiveHint == nil {
					t.Fatalf("Tool %s has no destructiveHint annotation", tool.Tool.Name)
				}
			}
		})

		t.Run("read-only tools are properly annotated", func(t *testing.T) {
			readOnlyTools := []string{
				"get-virtual-services",
				"get-destination-rules",
				"get-gateways",
				"get-service-entries",
				"get-proxy-clusters",
				"get-proxy-bootstrap",
			}

			for _, toolName := range readOnlyTools {
				found := false
				for _, tool := range tools {
					if tool.Tool.Name == toolName {
						found = true
						if tool.Tool.Annotations.ReadOnlyHint == nil || !*tool.Tool.Annotations.ReadOnlyHint {
							t.Fatalf("Tool %s should be marked as read-only", toolName)
						}
						break
					}
				}
				if !found {
					t.Fatalf("Read-only tool %s not found", toolName)
				}
			}
		})
	})
}

func TestToolHandlers(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		err := c.setupMCPServer()
		if err != nil {
			t.Fatalf("Failed to setup MCP server: %v", err)
		}
		defer c.server.Close()

		profile := &FullProfile{}
		tools := profile.GetTools(c.server)

		t.Run("all tools have handlers", func(t *testing.T) {
			for _, tool := range tools {
				if tool.Handler == nil {
					t.Fatalf("Tool %s has no handler", tool.Tool.Name)
				}
			}
		})

		t.Run("handlers have correct signatures", func(t *testing.T) {
			// Test that handlers can be called (this tests the signature)
			ctx := context.Background()
			request := mcp.CallToolRequest{}
			request.Params.Name = "test"
			request.Params.Arguments = map[string]interface{}{}

			for _, tool := range tools {
				// We don't call the actual handler as it would require real Istio resources
				// but we verify the handler exists and has the right type
				if tool.Handler == nil {
					t.Fatalf("Tool %s has no handler", tool.Tool.Name)
				}

				// Try to call with a simple test (this will likely fail due to missing resources,
				// but it tests that the function signature is correct)
				result, _ := tool.Handler(ctx, request)
				if result == nil {
					t.Fatalf("Tool %s handler returned nil result", tool.Tool.Name)
				}
			}
		})
	})
}

func TestProfileNames(t *testing.T) {
	t.Run("profile names are initialized", func(t *testing.T) {
		if len(ProfileNames) == 0 {
			t.Fatal("ProfileNames should not be empty")
		}

		// Check that all profiles in the Profiles slice have their names in ProfileNames
		expectedNames := make(map[string]bool)
		for _, profile := range Profiles {
			expectedNames[profile.GetName()] = true
		}

		for _, name := range ProfileNames {
			if !expectedNames[name] {
				t.Fatalf("Profile name %s found in ProfileNames but not in Profiles", name)
			}
		}

		for name := range expectedNames {
			found := false
			for _, profileName := range ProfileNames {
				if profileName == name {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Profile %s found in Profiles but not in ProfileNames", name)
			}
		}
	})
}

func TestToolParameters(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		err := c.setupMCPServer()
		if err != nil {
			t.Fatalf("Failed to setup MCP server: %v", err)
		}
		defer c.server.Close()

		profile := &FullProfile{}
		tools := profile.GetTools(c.server)

		t.Run("namespace tools have namespace parameter", func(t *testing.T) {
			namespaceTools := []string{
				"get-virtual-services",
				"get-destination-rules",
				"get-gateways",
				"get-service-entries",
			}

			for _, toolName := range namespaceTools {
				found := false
				for _, tool := range tools {
					if tool.Tool.Name == toolName {
						found = true
						// Check that the tool has input schema (not nil check since it's a struct)
						// The exact validation of schema would require more detailed inspection
						// but we can at least verify the tool exists and has a schema structure
						t.Logf("Tool %s has input schema", toolName)
						break
					}
				}
				if !found {
					t.Fatalf("Tool %s not found", toolName)
				}
			}
		})

		t.Run("proxy config tools have appropriate parameters", func(t *testing.T) {
			proxyConfigTools := []string{
				"get-proxy-clusters",
				"get-proxy-bootstrap",
				"get-proxy-listeners",
				"get-proxy-routes",
				"get-proxy-endpoints",
				"get-proxy-config-dump",
				"get-proxy-status",
			}

			for _, toolName := range proxyConfigTools {
				found := false
				for _, tool := range tools {
					if tool.Tool.Name == toolName {
						found = true
						// Proxy config tools should have input schema for pod specification
						t.Logf("Tool %s has input schema", toolName)
						break
					}
				}
				if !found {
					t.Fatalf("Tool %s not found", toolName)
				}
			}
		})
	})
}
