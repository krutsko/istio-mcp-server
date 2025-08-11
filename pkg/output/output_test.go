package output

import (
	"encoding/json"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestFromString tests output formatter lookup by name
func TestFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Output
	}{
		{
			name:     "yaml output",
			input:    "yaml",
			expected: Yaml,
		},
		{
			name:     "table output",
			input:    "table",
			expected: Table,
		},
		{
			name:     "json output",
			input:    "json",
			expected: Json,
		},
		{
			name:     "invalid output",
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
			result := FromString(tt.input)
			if result != tt.expected {
				t.Fatalf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestYamlOutput tests YAML output formatter functionality
func TestYamlOutput(t *testing.T) {
	output := Yaml

	t.Run("has correct name", func(t *testing.T) {
		if output.GetName() != "yaml" {
			t.Fatalf("Expected name 'yaml', got '%s'", output.GetName())
		}
	})

	t.Run("AsTable returns false", func(t *testing.T) {
		if output.AsTable() {
			t.Fatal("YAML output should not request table format")
		}
	})

	t.Run("prints simple object", func(t *testing.T) {
		obj := map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
		}

		result, err := output.PrintObj(obj)
		if err != nil {
			t.Fatalf("Failed to print object: %v", err)
		}

		if !strings.Contains(result, "apiVersion: v1") {
			t.Fatal("Expected YAML to contain 'apiVersion: v1'")
		}
		if !strings.Contains(result, "kind: Pod") {
			t.Fatal("Expected YAML to contain 'kind: Pod'")
		}
		if !strings.Contains(result, "name: test-pod") {
			t.Fatal("Expected YAML to contain 'name: test-pod'")
		}
	})

	t.Run("handles nil object", func(t *testing.T) {
		result, err := output.PrintObj(nil)
		if err != nil {
			t.Fatalf("Failed to print nil object: %v", err)
		}
		if strings.TrimSpace(result) != "null" {
			t.Fatalf("Expected 'null', got '%s'", strings.TrimSpace(result))
		}
	})
}

// TestJsonOutput tests JSON output formatter functionality
func TestJsonOutput(t *testing.T) {
	output := Json

	t.Run("has correct name", func(t *testing.T) {
		if output.GetName() != "json" {
			t.Fatalf("Expected name 'json', got '%s'", output.GetName())
		}
	})

	t.Run("AsTable returns false", func(t *testing.T) {
		if output.AsTable() {
			t.Fatal("JSON output should not request table format")
		}
	})

	t.Run("prints simple object", func(t *testing.T) {
		obj := map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
		}

		result, err := output.PrintObj(obj)
		if err != nil {
			t.Fatalf("Failed to print object: %v", err)
		}

		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}

		if parsed["apiVersion"] != "v1" {
			t.Fatal("Expected apiVersion to be 'v1'")
		}
		if parsed["kind"] != "Pod" {
			t.Fatal("Expected kind to be 'Pod'")
		}
	})

	t.Run("produces indented JSON", func(t *testing.T) {
		obj := map[string]interface{}{
			"test": "value",
		}

		result, err := output.PrintObj(obj)
		if err != nil {
			t.Fatalf("Failed to print object: %v", err)
		}

		// Check for indentation (should contain newlines and spaces)
		if !strings.Contains(result, "\n") {
			t.Fatal("Expected indented JSON to contain newlines")
		}
		if !strings.Contains(result, "  ") {
			t.Fatal("Expected indented JSON to contain spaces for indentation")
		}
	})

	t.Run("handles nil object", func(t *testing.T) {
		result, err := output.PrintObj(nil)
		if err != nil {
			t.Fatalf("Failed to print nil object: %v", err)
		}
		if strings.TrimSpace(result) != "null" {
			t.Fatalf("Expected 'null', got '%s'", result)
		}
	})
}

// TestTableOutput tests table output formatter functionality
func TestTableOutput(t *testing.T) {
	output := Table

	t.Run("has correct name", func(t *testing.T) {
		if output.GetName() != "table" {
			t.Fatalf("Expected name 'table', got '%s'", output.GetName())
		}
	})

	t.Run("AsTable returns true", func(t *testing.T) {
		if !output.AsTable() {
			t.Fatal("Table output should request table format")
		}
	})

	t.Run("prints object with table format message", func(t *testing.T) {
		obj := map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
		}

		result, err := output.PrintObj(obj)
		if err != nil {
			t.Fatalf("Failed to print object: %v", err)
		}

		// Since table format is not fully implemented, it should show a message
		if !strings.Contains(result, "Table format not fully implemented") {
			t.Fatal("Expected table output to mention it's not fully implemented")
		}
		if !strings.Contains(result, "showing YAML") {
			t.Fatal("Expected table output to mention it's showing YAML")
		}
	})
}

// TestOutputNames tests the Names slice and its relationship with Outputs
func TestOutputNames(t *testing.T) {
	t.Run("Names slice is initialized", func(t *testing.T) {
		if len(Names) == 0 {
			t.Fatal("Names slice should not be empty")
		}
	})

	t.Run("Names contains all output names", func(t *testing.T) {
		expectedNames := []string{"yaml", "table", "json"}

		if len(Names) != len(expectedNames) {
			t.Fatalf("Expected %d names, got %d", len(expectedNames), len(Names))
		}

		nameMap := make(map[string]bool)
		for _, name := range Names {
			nameMap[name] = true
		}

		for _, expected := range expectedNames {
			if !nameMap[expected] {
				t.Fatalf("Expected name '%s' not found in Names", expected)
			}
		}
	})

	t.Run("Names matches Outputs slice", func(t *testing.T) {
		if len(Names) != len(Outputs) {
			t.Fatalf("Names length (%d) doesn't match Outputs length (%d)", len(Names), len(Outputs))
		}

		outputNames := make(map[string]bool)
		for _, output := range Outputs {
			outputNames[output.GetName()] = true
		}

		for _, name := range Names {
			if !outputNames[name] {
				t.Fatalf("Name '%s' found in Names but corresponding output not found in Outputs", name)
			}
		}
	})
}

// TestOutputsSlice tests the Outputs slice
func TestOutputsSlice(t *testing.T) {
	t.Run("contains all expected outputs", func(t *testing.T) {
		expectedOutputs := []Output{Yaml, Table, Json}

		if len(Outputs) != len(expectedOutputs) {
			t.Fatalf("Expected %d outputs, got %d", len(expectedOutputs), len(Outputs))
		}

		for i, expected := range expectedOutputs {
			if Outputs[i] != expected {
				t.Fatalf("Expected output at index %d to be %v, got %v", i, expected, Outputs[i])
			}
		}
	})
}

// TestComplexObjects tests printing complex Kubernetes-like objects
func TestComplexObjects(t *testing.T) {
	// Test with a more complex Kubernetes-like object
	var podList unstructured.UnstructuredList
	err := json.Unmarshal([]byte(`{
		"apiVersion": "v1",
		"kind": "PodList",
		"items": [{
			"apiVersion": "v1",
			"kind": "Pod",
			"metadata": {
				"name": "test-pod",
				"namespace": "default",
				"labels": {
					"app": "test"
				}
			},
			"spec": {
				"containers": [{
					"name": "test-container",
					"image": "nginx"
				}]
			}
		}]
	}`), &podList)
	if err != nil {
		t.Fatalf("Failed to create test object: %v", err)
	}

	t.Run("YAML output handles complex objects", func(t *testing.T) {
		result, err := Yaml.PrintObj(&podList)
		if err != nil {
			t.Fatalf("Failed to print complex object as YAML: %v", err)
		}
		if !strings.Contains(result, "test-pod") {
			t.Fatal("Expected YAML to contain pod name")
		}
		if !strings.Contains(result, "nginx") {
			t.Fatal("Expected YAML to contain container image")
		}
	})

	t.Run("JSON output handles complex objects", func(t *testing.T) {
		result, err := Json.PrintObj(&podList)
		if err != nil {
			t.Fatalf("Failed to print complex object as JSON: %v", err)
		}

		// Verify it's valid JSON
		var parsed interface{}
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}

		if !strings.Contains(result, "test-pod") {
			t.Fatal("Expected JSON to contain pod name")
		}
		if !strings.Contains(result, "nginx") {
			t.Fatal("Expected JSON to contain container image")
		}
	})

	t.Run("Table output handles complex objects", func(t *testing.T) {
		result, err := Table.PrintObj(&podList)
		if err != nil {
			t.Fatalf("Failed to print complex object as table: %v", err)
		}
		if !strings.Contains(result, "Table format not fully implemented") {
			t.Fatal("Expected table output to show not implemented message")
		}
	})
}
