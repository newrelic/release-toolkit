package common

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ExcludedDependencies represents the structure of the YAML file.
type ExcludedDependencies struct {
	Dependencies []string `yaml:"dependencies"`
}

// LoadExcludedDependencies loads the YAML file and returns a slice of excluded dev dependencies.
func LoadExcludedDependencies(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}

	var excludedDependencies ExcludedDependencies
	err = yaml.Unmarshal(data, &excludedDependencies)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling YAML data: %w", err)
	}

	return excludedDependencies.Dependencies, nil
}
