package common

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadExcludedDependencies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		filePath string
		expected []string
		wantErr  bool
	}{
		{
			name:     "Test with valid file",
			filePath: path.Join("..", "testdata", "excluded-dependencies.yml"),
			expected: []string{"github.com/stretchr/testify", "github.com/testcontainers/testcontainers-go"},
			wantErr:  false,
		},
		{
			name:     "Test with non-existent file",
			filePath: path.Join("..", "testdata", "unexistent-file.yml"),
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := LoadExcludedDependencies(tc.filePath)

			if (err != nil) != tc.wantErr {
				t.Fatalf("loadExcludedDependencies error = %v, wantErr %v", err, tc.wantErr)
			}

			assert.Equal(t, tc.expected, got)
		})
	}
}
