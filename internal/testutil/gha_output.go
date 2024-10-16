package testutil

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/newrelic/release-toolkit/src/app/gha"
)

type GithubOutputWriter struct {
	File *os.File
}

func NewGithubOutputWriter(t *testing.T) GithubOutputWriter {
	t.Helper()

	ghaOutputFileName := path.Join(t.TempDir(), "temporary_github_output_file")
	t.Setenv(gha.GithubOutput, ghaOutputFileName)

	ghaOutputFile, err := os.Create(ghaOutputFileName)
	if err != nil {
		t.Fatalf("Error creating temporary GHA output file for test: %v", err)
	}

	return GithubOutputWriter{
		File: ghaOutputFile,
	}
}

func (ghaOut GithubOutputWriter) Result(t *testing.T) string {
	t.Helper()

	actual, err := io.ReadAll(ghaOut.File)
	if err != nil {
		t.Fatalf("Unable to read temporary GHA output file: %v", err)
	}
	return string(actual)
}
