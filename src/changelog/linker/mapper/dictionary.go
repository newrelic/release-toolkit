package mapper

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/newrelic/release-toolkit/src/changelog"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var ErrNilTemplateField = errors.New("required link template field is nil")

// Dictionary stores a mapping between dependency names and templates to obtain a changelog link.
// Templates have access to fields of the changelog.Dependency object.
type Dictionary struct {
	Changelogs map[string]string `yaml:"dictionary"`
}

func NewDictionary(r io.Reader) (Dictionary, error) {
	d := Dictionary{}

	buf, err := io.ReadAll(r)
	if err != nil {
		return d, fmt.Errorf("reading dictionary from source: %w", err)
	}

	err = yaml.Unmarshal(buf, &d)
	if err != nil {
		return d, fmt.Errorf("unmarshaling yaml: %w", err)
	}

	return d, nil
}

func (d Dictionary) Map(dep changelog.Dependency) string {
	if len(d.Changelogs) == 0 {
		return ""
	}

	if tpl := d.Changelogs[dep.Name]; tpl != "" {
		link, err := d.template(dep, tpl)
		if err != nil {
			log.Errorf("Error mapping changelog for %q from dictionary: %v", dep.Name, err)
			return ""
		}

		return link
	}

	log.Debugf("Dependency %q did not match any entry in the dictionary, attempting a partial match", dep.Name)
	for name, tpl := range d.Changelogs {
		if !strings.Contains(dep.Name, name) {
			continue
		}

		log.Infof("Linking changelog for %q from dictionary entry %q", dep.Name, name)

		link, err := d.template(dep, tpl)
		if err != nil {
			log.Errorf("Error mapping changelog for %q from dictionary: %v", dep.Name, err)
			return ""
		}

		return link
	}

	return ""
}

func (d Dictionary) template(dep changelog.Dependency, tplString string) (string, error) {
	tpl, err := template.New("changelog").Parse(tplString)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	linkTemplate := &strings.Builder{}
	err = tpl.Execute(linkTemplate, dep)
	if err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	link := linkTemplate.String()
	if strings.Contains(link, "<nil>") {
		return "", ErrNilTemplateField
	}

	return link, nil
}
