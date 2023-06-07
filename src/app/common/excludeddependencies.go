package common

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// excludedDependencies represents the structure of the YAML file.
type excludedDependencies struct {
	List []string `yaml:"dependencies"`
}

// LoadExcludedDependencies loads the YAML file and returns a slice of excluded dev dependencies.
func LoadExcludedDependencies(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}

	var excludedDeps excludedDependencies
	err = yaml.Unmarshal(data, &excludedDeps)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling YAML data: %w", err)
	}

	return excludedDeps.List, nil
}
