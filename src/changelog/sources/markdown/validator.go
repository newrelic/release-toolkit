package markdown

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown/headingdoc"
)

const (
	ChangelogHeader = "Changelog"
	LevelFirst      = iota
	LevelSecond
	LevelThird
)

// Validator is an object that validates a headingdoc.Doc.
type Validator struct {
	doc *headingdoc.Doc
}

var (
	ErrNoChangelogHeader      = errors.New("no Changelog top-header found")
	ErrEmptyHeldHeader        = errors.New("held header found with empty content")
	ErrNoUnreleasedL2Header   = errors.New("changelog must contain a L2 Unreleased header, even if it is empty")
	ErrL2WrongChildren        = errors.New("header can only contain L3 headers")
	ErrL3HeaderEmptyContent   = errors.New("header found with empty content")
	ErrL3HeaderNoItemizedList = errors.New("header must contain only an itemized list")
)

// NewValidator returns a new Validator for a headingdoc.Doc read from the supplied reader.
func NewValidator(r io.Reader) (Validator, error) {
	doc, err := headingdoc.NewFromReader(r)
	if err != nil {
		return Validator{}, fmt.Errorf("parsing markdown: %w", err)
	}
	return Validator{doc: doc}, nil
}

// Validate ensures the changelog has a correct format.
func (v *Validator) Validate() []error {
	errs := make([]error, 0)
	// Ensure one L1 header with the exact, case-sensitive check Changelog.
	if v.doc.Level != LevelFirst || v.doc.Name != ChangelogHeader {
		errs = append(errs, ErrNoChangelogHeader)
	}

	// Ensure one L2 header with the exact, case-sensitive check Unreleased.
	unreleased := v.doc.FindOne(unreleasedHeader)
	if unreleased == nil || unreleased.Level != LevelSecond {
		errs = append(errs, ErrNoUnreleasedL2Header)
	}

	for _, header := range v.doc.Children {
		errs = append(errs, v.checkL2Header(header)...)
	}

	return errs
}

func (v *Validator) checkL2Header(header *headingdoc.Doc) []error {
	errs := make([]error, 0)
	switch strings.ToLower(header.Name) {
	case unreleasedHeader:
		errs = append(errs, v.validateL2Children(header)...)
	case heldHeader:
		if len(header.Content) <= 1 {
			errs = append(errs, ErrEmptyHeldHeader)
		}
	default:
	}

	return errs
}

// validateL2Children validates L3 headers, that must not be empty and if they match a defined changelog.EntryType
// they must be an itemized list.
func (v *Validator) validateL2Children(l2doc *headingdoc.Doc) []error {
	errs := make([]error, 0)
	for _, header := range l2doc.Children {
		if header.Level != LevelThird {
			errs = append(errs, fmt.Errorf("%q %w", l2doc.Name, ErrL2WrongChildren))
		}

		if len(header.Content) <= 1 {
			errs = append(errs, fmt.Errorf("%q %w", header.Name, ErrL3HeaderEmptyContent))
			continue
		}

		errs = append(errs, v.ensureEntryTypeItemizedList(header)...)
	}

	return errs
}

// ensureEntryTypeItemizedList ensures L3 headers that match one of the defined changelog.EntryType
// must contain only an itemized list.
func (v *Validator) ensureEntryTypeItemizedList(header *headingdoc.Doc) []error {
	errs := make([]error, 0)
	if strings.ToLower(header.Name) == string(changelog.TypeEnhancement) ||
		strings.ToLower(header.Name) == string(changelog.TypeBugfix) ||
		strings.ToLower(header.Name) == string(changelog.TypeSecurity) ||
		strings.ToLower(header.Name) == string(changelog.TypeBreaking) ||
		strings.ToLower(header.Name) == string(changelog.TypeDependency) {
		switch header.Content[1].(type) {
		default:
			errs = append(errs, fmt.Errorf("%q %w", header.Name, ErrL3HeaderNoItemizedList))
		case *ast.List:
		}
	}

	return errs
}
