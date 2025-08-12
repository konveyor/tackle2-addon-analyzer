package builder

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	liberr "github.com/jortel/go-utils/error"
	output "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/pointer"
)

var (
	addon = hub.Addon
	wrap  = liberr.Wrap
)

// NewInsights returns a new insights builder.
func NewInsights(path string) (b *Insights, err error) {
	b = &Insights{}
	err = b.read(path)
	return
}

// Insights builds insights and facts.
type Insights struct {
	ruleErr RuleError
	facts   []api.Fact
	input   []output.RuleSet
}

// RuleError returns the rule error.
func (b *Insights) RuleError() (r *RuleError) {
	for _, ruleset := range b.input {
		b.ruleErr.Append(ruleset)
	}
	return &b.ruleErr
}

// Write insights section.
func (b *Insights) Write(writer io.Writer) (err error) {
	b.ensureUnique()
	_, _ = writer.Write([]byte(api.BeginInsightsMarker))
	_, _ = writer.Write([]byte{'\n'})
	for _, ruleset := range b.input {
		for ruleid, v := range ruleset.Violations {
			insight := api.Insight{
				RuleSet:     ruleset.Name,
				Rule:        ruleid,
				Description: v.Description,
				Labels:      v.Labels,
			}
			if v.Category != nil {
				insight.Category = string(*v.Category)
			}
			if v.Effort != nil {
				insight.Effort = *v.Effort
			}
			insight.Links = []api.Link{}
			for _, l := range v.Links {
				insight.Links = append(
					insight.Links,
					api.Link{
						URL:   l.URL,
						Title: l.Title,
					})
			}
			insight.Incidents = []api.Incident{}
			for _, i := range v.Incidents {
				incident := api.Incident{
					File:     b.fileRef(i.URI),
					Line:     pointer.IntDeref(i.LineNumber, 0),
					Message:  i.Message,
					CodeSnip: i.CodeSnip,
					Facts:    i.Variables,
				}
				insight.Incidents = append(
					insight.Incidents,
					incident)
			}
			err = b.encode(writer, &insight)
			if err != nil {
				return
			}
		}
		for ruleid, v := range ruleset.Insights {
			insight := api.Insight{
				RuleSet:     ruleset.Name,
				Rule:        ruleid,
				Description: v.Description,
				Labels:      v.Labels,
			}
			insight.Links = []api.Link{}
			for _, l := range v.Links {
				insight.Links = append(
					insight.Links,
					api.Link{
						URL:   l.URL,
						Title: l.Title,
					})
			}
			insight.Incidents = []api.Incident{}
			for _, i := range v.Incidents {
				incident := api.Incident{
					File:     b.fileRef(i.URI),
					Line:     pointer.IntDeref(i.LineNumber, 0),
					Message:  i.Message,
					CodeSnip: i.CodeSnip,
					Facts:    i.Variables,
				}
				insight.Incidents = append(
					insight.Incidents,
					incident)
			}
			err = b.encode(writer, &insight)
			if err != nil {
				err = wrap(err)
				return
			}
		}
	}
	_, _ = writer.Write([]byte(api.EndInsightsMarker))
	_, _ = writer.Write([]byte{'\n'})
	return
}

// encode object.
func (b *Insights) encode(writer io.Writer, r any) (err error) {
	encoder := yaml.NewEncoder(writer)
	err = encoder.Encode(r)
	if err != nil {
		return
	}
	err = encoder.Close()
	return
}

// read ruleSets.
func (b *Insights) read(path string) (err error) {
	b.input = []output.RuleSet{}
	f, err := os.Open(path)
	if err != nil {
		err = wrap(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	d := yaml.NewDecoder(f)
	err = d.Decode(&b.input)
	err = wrap(err)
	return
}

// fileRef returns the file (relative) path.
func (b *Insights) fileRef(in uri.URI) (s string) {
	s = string(in)
	u, err := url.Parse(s)
	if err == nil {
		s = u.Path
	}
	return
}

// Tags builds tags.
func (b *Insights) Tags() (tags []string) {
	for _, r := range b.input {
		tags = append(tags, r.Tags...)
	}
	return
}

// Facts builds facts.
func (b *Insights) Facts() (facts api.Map) {
	return
}

// ensureUnique detect rules reporting both violation and insight.
// Append (_) suffix to ruleid as needed.
func (b *Insights) ensureUnique() {
	rules := make(map[string]int8)
	for _, ruleset := range b.input {
		collections := []map[string]output.Violation{
			ruleset.Violations,
			ruleset.Insights,
		}
		for _, violations := range collections {
			for ruleid, v := range violations {
				key := ruleset.Name + ruleid
				if _, found := rules[key]; found {
					delete(violations, ruleid)
					ruleid += "_"
					violations[ruleid] = v
				}
				rules[key]++
			}
		}
	}
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
	if len(e.items) == 0 {
		return
	}
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
