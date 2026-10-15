package changes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/newrelic/release-toolkit/src/changelog"
	"gopkg.in/yaml.v3"
)

const DefaultDir = ".changelog"

var (
	ErrInvalidChanges   = errors.New("invalid change files")
	ErrEmptyChange      = errors.New("change file is empty")
	ErrNoType           = errors.New("type is required")
	ErrUnknownType      = errors.New("unknown type")
	ErrNoMessage        = errors.New("message is required")
	ErrMultilineMessage = errors.New("message must be a single line")
)

type Change struct {
	Type    string `yaml:"type"`
	Message string `yaml:"message"`
}

type Source struct {
	dir string
}

func NewSource(dir string) Source {
	return Source{dir: dir}
}

func (s Source) Files() ([]string, error) {
	dirEntries, err := os.ReadDir(s.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading change files directory %q: %w", s.dir, err)
	}

	files := make([]string, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || !isChangeFile(dirEntry.Name()) {
			continue
		}

		files = append(files, filepath.Join(s.dir, dirEntry.Name()))
	}

	sort.Strings(files)

	return files, nil
}

func isChangeFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".yaml" && ext != ".yml" {
		return false
	}

	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}

	return !strings.EqualFold(strings.TrimSuffix(name, filepath.Ext(name)), "template")
}

func (s Source) Changelog() (*changelog.Changelog, error) {
	ch := &changelog.Changelog{}

	files, err := s.Files()
	if err != nil {
		return nil, err
	}

	var problems []string
	for _, file := range files {
		entry, loadErr := Load(file)
		if loadErr != nil {
			problems = append(problems, loadErr.Error())
			continue
		}

		ch.Changes = append(ch.Changes, entry)
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidChanges, strings.Join(problems, "; "))
	}

	return ch, nil
}

func (s Source) Validate() []error {
	files, err := s.Files()
	if err != nil {
		return []error{err}
	}

	var errs []error
	for _, file := range files {
		if _, loadErr := Load(file); loadErr != nil {
			errs = append(errs, loadErr)
		}
	}

	return errs
}

func (s Source) Remove() ([]string, error) {
	files, err := s.Files()
	if err != nil {
		return nil, err
	}

	removed := make([]string, 0, len(files))
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return removed, fmt.Errorf("removing change file: %w", err)
		}

		removed = append(removed, file)
	}

	return removed, nil
}

func Load(path string) (changelog.Entry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return changelog.Entry{}, fmt.Errorf("%s: reading change file: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)

	change := Change{}
	if err = decoder.Decode(&change); err != nil {
		if errors.Is(err, io.EOF) {
			return changelog.Entry{}, fmt.Errorf("%s: %w", path, ErrEmptyChange)
		}

		return changelog.Entry{}, fmt.Errorf("%s: parsing change file: %w", path, err)
	}

	entry, err := change.Entry()
	if err != nil {
		return changelog.Entry{}, fmt.Errorf("%s: %w", path, err)
	}

	return entry, nil
}

func (c Change) Entry() (changelog.Entry, error) {
	if strings.TrimSpace(c.Type) == "" {
		return changelog.Entry{}, ErrNoType
	}

	entryType, isEntryType := changelog.EntryTypeFromTitle(c.Type)
	if !isEntryType {
		return changelog.Entry{}, fmt.Errorf("%w %q, use one of: %s", ErrUnknownType, c.Type, entryTypes())
	}

	message := strings.TrimSpace(c.Message)
	if message == "" {
		return changelog.Entry{}, ErrNoMessage
	}

	if strings.ContainsAny(message, "\r\n") {
		return changelog.Entry{}, ErrMultilineMessage
	}

	return changelog.Entry{
		Type:    entryType,
		Message: message,
	}, nil
}

func entryTypes() string {
	types := make([]string, 0, len(changelog.EntryTypes()))
	for _, entryType := range changelog.EntryTypes() {
		types = append(types, string(entryType))
	}

	return strings.Join(types, ", ")
}
