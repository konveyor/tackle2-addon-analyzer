package builder

import (
	"encoding/json"
	output "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	"github.com/konveyor/tackle2-hub/api"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

//
// Issues builds issues and facts.
type Issues struct {
	facts []api.Fact
	Path  string
}

//
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

//
// Write issues to the writer.
func (b *Issues) Write(writer io.Writer) (err error) {
	input, err := b.read()
	if err != nil {
		return
	}
	encoder := yaml.NewEncoder(writer)
	for _, ruleset := range input {
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
					File:     b.uriStr(i.URI),
					Message:  i.Message,
					CodeSnip: i.CodeSnip,
					Facts:    i.Variables,
				}
				issue.Incidents = append(
					issue.Incidents,
					incident)
			}
			_ = encoder.Encode(&issue)
		}
	}
	return
}

//
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

//
// uniStr (safely) returns URI filename.
func (b *Issues) uriStr(in uri.URI) string {
	defer func() {
		recover()
	}()
	return in.Filename()
}

//
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

//
// Facts builds facts.
func (b *Issues) Facts() (facts []api.Fact) {
	input, err := b.read()
	if err != nil {
		return
	}
	for _, r := range input {
		for _, v := range r.Violations {
			mp := make(map[string]interface{})
			_ = json.Unmarshal(v.Extras, &mp)
			for k, v := range mp {
				facts = append(
					facts,
					api.Fact{
						Key:   k,
						Value: v,
					})
			}
		}
	}
	return
}
