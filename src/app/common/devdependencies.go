package common

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DevDependencies represents the structure of the YAML file.
type DevDependencies struct {
	ExcludedDevDependencies []string `yaml:"excluded_dev_dependencies"`
}

// LoadExcludedDevDependencies loads the YAML file and returns a slice of excluded dev dependencies.
func LoadExcludedDevDependencies(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}

	var devDependencies DevDependencies
	err = yaml.Unmarshal(data, &devDependencies)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling YAML data: %w", err)
	}

	return devDependencies.ExcludedDevDependencies, nil
}
