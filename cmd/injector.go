package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
)

// KeyRegex $(variable)
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

// TypeError used to report resource field cast error.
type TypeError struct {
	Field  *Field
	Reason string
	Object any
}

func (e *TypeError) Error() (s string) {
	return fmt.Sprintf(
		"Resource injector: cast failed. field=%s type=%s reason=%s, object:%v",
		e.Field.Name,
		e.Field.Type,
		e.Reason,
		e.Object)
}

func (e *TypeError) Is(err error) (matched bool) {
	var inst *TypeError
	matched = errors.As(err, &inst)
	return
}

// KeyConflictError reports key redefined errors.
type KeyConflictError struct {
	Key   string
	Value any
}

func (e *KeyConflictError) Error() (s string) {
	return fmt.Sprintf(
		"Key: '%s' = '%v' cannot be redefined.",
		e.Key,
		e.Value)
}

func (e *KeyConflictError) Is(err error) (matched bool) {
	var inst *KeyConflictError
	matched = errors.As(err, &inst)
	return
}

// Field injection specification.
type Field struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Key     string `json:"key"`
	Type    string `json:"type"`
	Default any    `json:"default"`
}

// cast returns object cast as defined by Field.Type.
func (f *Field) cast(object any) (cast any, err error) {
	cast = object
	if f.Type == "" {
		return
	}
	defer func() {
		if err != nil {
			err = &TypeError{
				Field:  f,
				Reason: err.Error(),
				Object: object,
			}
		}
	}()
	switch strings.ToLower(f.Type) {
	case "string":
		cast = fmt.Sprintf("%v", object)
	case "integer":
		switch x := object.(type) {
		case int,
			int8,
			int16,
			int32,
			int64:
			cast = x
		case bool:
			cast = 0
			if x {
				cast = 1
			}
		case string:
			cast, err = strconv.Atoi(x)
		default:
			err = errors.New("expected: integer|boolean|string")
		}
	case "boolean":
		switch x := object.(type) {
		case bool:
			cast = x
		case int,
			int8,
			int16,
			int32,
			int64:
			cast = x != 0
		case string:
			cast, err = strconv.ParseBool(x)
		default:
			err = errors.New("expected: integer|boolean|string")
		}
	default:
		err = errors.New("expected: integer|boolean|string")
	}
	return
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

// Injector replaces variables in the object.
// format: $(variable).
type Injector struct {
	dict map[string]any
}

// Inject resources into extension metadata.
func (r *Injector) Inject(md *Metadata) (err error) {
	r.init()
	mp := r.asMap(md)
	mp = r.inject(mp).(map[string]any)
	err = r.object(mp, md)
	if err != nil {
		return
	}
	return
}

// Use map.
func (r *Injector) Use(d map[string]any) {
	r.dict = d
}

// constructor.
func (r *Injector) init() {
	if r.dict == nil {
		r.dict = make(map[string]any)
	}
}

// inject replaces `dict` variables referenced in metadata.
func (r *Injector) inject(in any) (out any) {
	if r.dict == nil {
		return
	}
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
			v := r.dict[match[2]]
			if len(node) > len(match[0]) {
				node = strings.Replace(
					node,
					match[0],
					r.string(v),
					-1)
			} else {
				out = v
				return
			}
		}
		out = node
	default:
		out = node
	}
	return
}

// objectMap returns a map for a resource object.
func (r *Injector) asMap(object any) (mp map[string]any) {
	b, _ := json.Marshal(object)
	mp = make(map[string]any)
	_ = json.Unmarshal(b, &mp)
	return
}

// objectMap returns a map for a resource object.
func (r *Injector) object(mp map[string]any, object any) (err error) {
	b, _ := json.Marshal(mp)
	err = json.Unmarshal(b, object)
	return
}

// string returns a string representation of a field value.
func (r *Injector) string(object any) (s string) {
	if object != nil {
		s = fmt.Sprintf("%v", object)
	}
	return
}

// ResourceInjector inject resources into extension metadata.
// Example:
//
//	metadata:
//	 provider:
//	   address: localhost:$(PORT)
//	   initConfig:
//	   - providerSpecificConfig:
//	       mavenInsecure: $(maven.insecure)
//	       mavenSettingsFile: $(maven.settings.path)
//	   name: java
//	 resources:
//	 - selector: identity:kind=maven
//	   fields:
//	   - key: maven.settings.path
//	     name: settings
//	     path: /shared/creds/maven/settings.xml
//	 - selector: setting:key=mvn.insecure.enabled
//	   fields:
//	   - key: maven.insecure
//	     name: value
type ResourceInjector struct {
	Injector
}

// Inject resources into extension metadata.
func (r *ResourceInjector) Inject(md *Metadata) (err error) {
	r.init()
	err = r.build(md)
	if err != nil {
		return
	}
	err = r.Injector.Inject(md)
	return
}

// build builds resource dictionary.
func (r *ResourceInjector) build(md *Metadata) (err error) {
	application, err := addon.Task.Application()
	if err != nil {
		return
	}
	for _, resource := range md.Resources {
		err = r.addDefaults(&resource)
	}
	for _, resource := range md.Resources {
		parsed := ParsedSelector{}
		parsed.With(resource.Selector)
		switch strings.ToLower(parsed.kind) {
		case "identity":
			identity, found, nErr :=
				addon.Application.Identity(application.ID).Search().
					Direct(parsed.value).
					Indirect(parsed.value).
					Find()
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
		case "setting":
			setting := &api.Setting{}
			err = addon.Setting.Get(parsed.value, &setting.Value)
			if err != nil {
				return
			}
			err = r.add(&resource, setting)
			if err != nil {
				return
			}
		default:
			err = &SelectorNotSupported{Selector: resource.Selector}
			return
		}
	}
	return
}

// addDefaults adds defaults when specified.
func (r *ResourceInjector) addDefaults(resource *Resource) (err error) {
	for _, f := range resource.Fields {
		if f.Default == nil {
			continue
		}
		err = r.addField(&f, f.Default)
		if err != nil {
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
		err = r.addField(&f, v)
		if err != nil {
			return
		}
	}
	return
}

// addField adds field to the dict.
// When field has a path defined, the values is written to the
// file and the dict[key] = path.
func (r *ResourceInjector) addField(f *Field, v any) (err error) {
	if f.Path != "" {
		err = r.write(f.Path, v)
		if err != nil {
			return
		}
		v = f.Path
	} else {
		v, err = f.cast(v)
		if err != nil {
			return
		}
	}
	if _, found := r.dict[f.Key]; found {
		err = &KeyConflictError{
			Key:   f.Key,
			Value: v,
		}
		return
	}
	r.dict[f.Key] = v
	return
}

// write a resource field value to a file.
func (r *ResourceInjector) write(path string, object any) (err error) {
	err = nas.MkDir(filepath.Dir(path), 0755)
	if err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	s := r.string(object)
	_, err = f.Write([]byte(s))
	return
}
