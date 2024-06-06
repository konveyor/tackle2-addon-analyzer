package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/konveyor/tackle2-hub/api"
)

var (
	KeyRegex = regexp.MustCompile(`(\$\()([^)]+)(\))`)
)

// SelectorNotSupported used to report not supported.
type SelectorNotSupported struct {
	Selector string
}

func (e *SelectorNotSupported) Error() (s string) {
	return fmt.Sprintf("Resource selector='%s', not-supported.", e.Selector)
}

func (e *SelectorNotSupported) Is(err error) (matched bool) {
	var inst *SelectorNotSupported
	matched = errors.As(err, &inst)
	return
}

// FieldNotMatched used to report resource field not matched.
type FieldNotMatched struct {
	Kind  string
	Field string
}

func (e *FieldNotMatched) Error() (s string) {
	return fmt.Sprintf("Resource injector: field=%s.%s, not-matched.", e.Kind, e.Field)
}

func (e *FieldNotMatched) Is(err error) (matched bool) {
	var inst *FieldNotMatched
	matched = errors.As(err, &inst)
	return
}

// Field injection specification.
type Field struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Key  string `json:"key"`
}

// Resource injection specification.
// Format: <kind>:<key>=<value>
type Resource struct {
	Selector string  `json:"selector"`
	Fields   []Field `json:"fields"`
}

// Metadata for provider extensions.
type Metadata struct {
	Resources []Resource      `json:"resources,omitempty"`
	Provider  provider.Config `json:"provider"`
}

// ParsedSelector -
type ParsedSelector struct {
	ns    string
	kind  string
	name  string
	value string
}

// With parses and populates the selector.
func (p *ParsedSelector) With(s string) {
	part := strings.SplitN(s, "/", 2)
	if len(part) > 1 {
		p.ns = part[0]
		s = part[1]
	}
	part = strings.SplitN(s, ":", 2)
	if len(part) > 1 {
		p.kind = part[0]
		s = part[1]
	}
	part = strings.SplitN(s, "=", 2)
	p.name = part[0]
	if len(part) > 1 {
		p.value = part[1]
	}
}

// ResourceInjector inject resources into extension metadata.
type ResourceInjector struct {
	dict map[string]string
}

// Inject resources into extension metadata.
// Returns injected provider (settings).
func (r *ResourceInjector) Inject(extension *api.Extension) (p *provider.Config, err error) {
	mp := r.asMap(extension.Metadata)
	md := Metadata{}
	err = r.object(mp, &md)
	if err != nil {
		return
	}
	err = r.build(&md)
	if err != nil {
		return
	}
	mp = r.asMap(&md.Provider)
	mp = r.inject(mp).(map[string]any)
	err = r.object(mp, &md.Provider)
	if err != nil {
		return
	}
	p = &md.Provider
	return
}

// build builds resource dictionary.
func (r *ResourceInjector) build(md *Metadata) (err error) {
	r.dict = make(map[string]string)
	application, err := addon.Task.Application()
	if err != nil {
		return
	}
	for _, resource := range md.Resources {
		parsed := ParsedSelector{}
		parsed.With(resource.Selector)
		switch strings.ToLower(parsed.kind) {
		case "identity":
			identity, found, nErr := addon.Application.FindIdentity(application.ID, parsed.value)
			if nErr != nil {
				err = nErr
				return
			}
			if found {
				err = r.add(&resource, identity)
				if err != nil {
					return
				}
			}
		default:
			err = &SelectorNotSupported{Selector: resource.Selector}
			return
		}
	}
	return
}

// add the resource fields specified in the injector.
func (r *ResourceInjector) add(resource *Resource, object any) (err error) {
	mp := r.asMap(object)
	for _, f := range resource.Fields {
		v, found := mp[f.Name]
		if !found {
			err = &FieldNotMatched{
				Kind:  resource.Selector,
				Field: f.Name,
			}
			return
		}
		fv := r.string(v)
		if f.Path != "" {
			err = r.write(f.Path, fv)
			if err != nil {
				return
			}
			fv = f.Path
		}
		r.dict[f.Key] = fv
	}
	return
}

// write a resource field value to a file.
func (r *ResourceInjector) write(path string, s string) (err error) {
	f, err := os.Create(path)
	if err == nil {
		_, err = f.Write([]byte(s))
		_ = f.Close()
	}
	return
}

// string returns a string representation of a field value.
func (r *ResourceInjector) string(object any) (s string) {
	if object != nil {
		s = fmt.Sprintf("%v", object)
	}
	return
}

// objectMap returns a map for a resource object.
func (r *ResourceInjector) asMap(object any) (mp map[string]any) {
	b, _ := json.Marshal(object)
	mp = make(map[string]any)
	_ = json.Unmarshal(b, &mp)
	return
}

// objectMap returns a map for a resource object.
func (r *ResourceInjector) object(mp map[string]any, object any) (err error) {
	b, _ := json.Marshal(mp)
	err = json.Unmarshal(b, object)
	return
}

// inject replaces `dict` variables referenced in metadata.
func (r *ResourceInjector) inject(in any) (out any) {
	switch node := in.(type) {
	case map[string]any:
		for k, v := range node {
			node[k] = r.inject(v)
		}
		out = node
	case []any:
		var injected []any
		for _, n := range node {
			injected = append(
				injected,
				r.inject(n))
		}
		out = injected
	case string:
		for {
			match := KeyRegex.FindStringSubmatch(node)
			if len(match) < 3 {
				break
			}
			node = strings.Replace(
				node,
				match[0],
				r.dict[match[2]],
				-1)
		}
		out = node
	default:
		out = node
	}
	return
}
