package main

import (
	"strings"

	"github.com/konveyor/tackle2-addon/command"
)

// Scope settings.
type Scope struct {
	WithKnownLibs bool `json:"withKnownLibs"`
	Packages      struct {
		Included []string `json:"included,omitempty"`
		Excluded []string `json:"excluded,omitempty"`
	} `json:"packages"`
}

// AddOptions adds analyzer options.
func (r *Scope) AddOptions(options *command.Options, mode Mode) (err error) {
	// If withDeps is false, we are only every doing source analysis
	// adding a dep label selector is strictly wrong in this situation
	if mode.WithDeps {
		// We want to filter out open source violations when we are not running
		// with known libraries.
		if !r.WithKnownLibs {
			options.Add(
				"--dep-label-selector",
				"!konveyor.io/dep-source=open-source")
		}

	}
	selector := r.incidentSelector()
	if selector != "" {
		options.Add("--incident-selector", selector)
	}
	return
}

// incidentSelector returns an incident selector.
// The injected `!package` matches incidents without a package variable.
func (r *Scope) incidentSelector() (selector string) {
	predicate := func(in []string) (p string) {
		var refs []string
		for _, s := range in {
			refs = append(refs, "package="+s)
		}
		p = strings.Join(refs, "||")
		return
	}
	var predicates []string
	p := predicate(r.Packages.Included)
	if len(p) > 0 {
		p = "(!package||" + p + ")"
		predicates = append(predicates, p)
	}
	p = predicate(r.Packages.Excluded)
	if len(p) > 0 {
		if len(predicates) == 0 {
			p = "!(package||" + p + ")"
		} else {
			p = "!(" + p + ")"
		}
		predicates = append(predicates, p)
	}
	selector = strings.Join(predicates, " && ")
	return
}
