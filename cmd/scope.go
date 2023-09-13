package main

import "github.com/konveyor/tackle2-addon/command"

//
// Scope settings.
type Scope struct {
	WithKnownLibs bool `json:"withKnownLibs"`
	Packages      struct {
		Included []string `json:"included,omitempty"`
		Excluded []string `json:"excluded,omitempty"`
	} `json:"packages"`
}

//
// AddOptions adds analyzer options.
func (r *Scope) AddOptions(options *command.Options) (err error) {
	if !r.WithKnownLibs {
		options.Add(
			"--dep-label-selector",
			"!konveyor.io/dep-source=open-source")
	}
	if len(r.Packages.Included) > 0 {
		options.Add("--packages", r.Packages.Included...)
	}
	if len(r.Packages.Excluded) > 0 {
		options.Add("--excludePackages", r.Packages.Excluded...)
	}
	return
}
