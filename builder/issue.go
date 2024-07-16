package builder

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	output "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/pointer"
)

var (
	addon = hub.Addon
)

// Issues builds issues and facts.
type Issues struct {
	ruleErr RuleError
	facts   []api.Fact
	Path    string
}

// Reader returns a reader.
func (b *Issues) Reader() (r io.Reader) {
	r, w := io.Pipe()
	go func() {
		var err error
		defer func() {
			if err != nil {
				_ = w.CloseWithError(err)
			} else {
				_ = w.Close()
			}
		}()
		err = b.Write(w)
	}()
	return
}

// Write issues to the writer.
func (b *Issues) Write(writer io.Writer) (err error) {
	input, err := b.read()
	if err != nil {
		return
	}
	encoder := yaml.NewEncoder(writer)
	for _, ruleset := range input {
		b.ruleErr.Append(ruleset)
		for ruleid, v := range ruleset.Violations {
			issue := api.Issue{
				RuleSet:     ruleset.Name,
				Rule:        ruleid,
				Description: v.Description,
				Labels:      v.Labels,
			}
			if v.Category != nil {
				issue.Category = string(*v.Category)
			}
			if v.Effort != nil {
				issue.Effort = *v.Effort
			}
			issue.Links = []api.Link{}
			for _, l := range v.Links {
				issue.Links = append(
					issue.Links,
					api.Link{
						URL:   l.URL,
						Title: l.Title,
					})
			}
			issue.Incidents = []api.Incident{}
			for _, i := range v.Incidents {
				incident := api.Incident{
					File:     b.fileRef(i.URI),
					Line:     pointer.IntDeref(i.LineNumber, 0),
					Message:  i.Message,
					CodeSnip: i.CodeSnip,
					Facts:    i.Variables,
				}
				issue.Incidents = append(
					issue.Incidents,
					incident)
			}
			err = encoder.Encode(&issue)
			if err != nil {
				return
			}
		}
	}
	if err != nil {
		return
	}
	if b.ruleErr.NotEmpty() {
		err = &b.ruleErr
		return
	}
	return
}

// read ruleSets.
func (b *Issues) read() (input []output.RuleSet, err error) {
	input = []output.RuleSet{}
	f, err := os.Open(b.Path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	bfr, err := io.ReadAll(f)
	err = yaml.Unmarshal(bfr, &input)
	return
}

// fileRef returns the file (relative) path.
func (b *Issues) fileRef(in uri.URI) (s string) {
	s = string(in)
	u, err := url.Parse(s)
	if err == nil {
		s = u.Path
	}
	return
}

// Tags builds tags.
func (b *Issues) Tags() (tags []string) {
	input, err := b.read()
	if err != nil {
		return
	}
	for _, r := range input {
		tags = append(tags, r.Tags...)
	}
	return
}

// Facts builds facts.
func (b *Issues) Facts() (facts api.FactMap) {
	return
}

// RuleError reported by the analyzer.
type RuleError struct {
	items map[string]string
}

func (e *RuleError) Error() (s string) {
	s = fmt.Sprintf(
		"Analyser reported %d errors.",
		len(e.items))
	return
}

func (e *RuleError) Is(err error) (matched bool) {
	var ruleError *RuleError
	matched = errors.As(err, &ruleError)
	return
}

func (e *RuleError) Append(ruleset output.RuleSet) {
	if e.items == nil {
		e.items = make(map[string]string)
	}
	for ruleid, err := range ruleset.Errors {
		ruleid := ruleset.Name + "." + ruleid
		e.items[ruleid] = err
	}
}

func (e *RuleError) NotEmpty() (b bool) {
	return len(e.items) > 0
}

func (e *RuleError) Report() {
	var errors []api.TaskError
	for ruleid, err := range e.items {
		errors = append(
			errors,
			api.TaskError{
				Severity:    "Error",
				Description: fmt.Sprintf("[Analyzer] %s: %s", ruleid, err),
			})
	}
	addon.Error(errors...)
}
