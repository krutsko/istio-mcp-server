package output

import (
	"encoding/json"
	"fmt"

	yml "sigs.k8s.io/yaml"
)

// Predefined output formatters
var Yaml = &yamlOutput{}
var Table = &tableOutput{}
var Json = &jsonOutput{}

// Output defines the interface for output formatters
type Output interface {
	// GetName returns the name of the output format, will be used by the CLI to identify the output format.
	GetName() string
	// AsTable true if the kubernetes request should be made with the `application/json;as=Table;v=0.1` header.
	AsTable() bool
	// PrintObj prints the given object as a string.
	PrintObj(obj interface{}) (string, error)
}

// Outputs contains all available output formatters
var Outputs = []Output{
	Yaml,
	Table,
	Json,
}

// Names contains the names of all available output formats
var Names []string

// FromString returns an output formatter by name
func FromString(name string) Output {
	for _, output := range Outputs {
		if output.GetName() == name {
			return output
		}
	}
	return nil
}

// yamlOutput provides YAML formatting
type yamlOutput struct{}

// GetName returns the output format name
func (p *yamlOutput) GetName() string {
	return "yaml"
}

// AsTable returns false for YAML output
func (p *yamlOutput) AsTable() bool {
	return false
}

// PrintObj formats the object as YAML
func (p *yamlOutput) PrintObj(obj interface{}) (string, error) {
	ret, err := yml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// tableOutput provides table formatting
type tableOutput struct{}

// GetName returns the output format name
func (p *tableOutput) GetName() string {
	return "table"
}

// AsTable returns true for table output
func (p *tableOutput) AsTable() bool {
	return true
}

// PrintObj formats the object as a table (currently falls back to YAML)
func (p *tableOutput) PrintObj(obj interface{}) (string, error) {
	// For now, just return YAML format since we don't have table printer
	// In a real implementation, this would use k8s.io/cli-runtime table printer
	yamlStr, err := Yaml.PrintObj(obj)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Table format not fully implemented, showing YAML:\n%s", yamlStr), nil
}

// jsonOutput provides JSON formatting
type jsonOutput struct{}

// GetName returns the output format name
func (p *jsonOutput) GetName() string {
	return "json"
}

// AsTable returns false for JSON output
func (p *jsonOutput) AsTable() bool {
	return false
}

// PrintObj formats the object as JSON
func (p *jsonOutput) PrintObj(obj interface{}) (string, error) {
	ret, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// init initializes the output format names
func init() {
	Names = make([]string, 0)
	for _, output := range Outputs {
		Names = append(Names, output.GetName())
	}
}
