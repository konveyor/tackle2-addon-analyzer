package main

import (
	"os"
	"path"

	"github.com/konveyor/tackle2-addon-analyzer/builder"
	"github.com/konveyor/tackle2-addon/command"
)

type RuleError = builder.RuleError

// Analyzer application analyzer.
type Analyzer struct {
	*Data
}

// Run analyzer.
func (r *Analyzer) Run() (issueBuilder *builder.Issues, depBuilder *builder.Deps, err error) {
	output := path.Join(Dir, "issues.yaml")
	depOutput := path.Join(Dir, "deps.yaml")
	cmd := command.New("/usr/local/bin/konveyor-analyzer")
	cmd.Options, err = r.options(output, depOutput)
	if err != nil {
		return
	}
	if Verbosity > 0 {
		cmd.Reporter.Verbosity = command.LiveOutput
	}
	issueBuilder = &builder.Issues{Path: output}
	depBuilder = &builder.Deps{Path: depOutput}
	err = cmd.Run()
	if err != nil {
		return
	}
	if Verbosity > 0 {
		f, pErr := addon.File.Post(output)
		if pErr != nil {
			err = pErr
			return
		}
		addon.Attach(f)
		if _, stErr := os.Stat(depOutput); stErr == nil {
			f, pErr = addon.File.Post(depOutput)
			if pErr != nil {
				err = pErr
				return
			}
			addon.Attach(f)
		}
	}
	return
}

// options builds Analyzer options.
func (r *Analyzer) options(output, depOutput string) (options command.Options, err error) {
	settings := &Settings{}
	err = settings.AppendExtensions()
	if err != nil {
		return
	}
	options = command.Options{
		"--provider-settings",
		settings.path(),
		"--output-file",
		output,
	}
	if !r.Data.Mode.Discovery {
		options.Add("--dep-output-file", depOutput)
	}
	err = r.Tagger.AddOptions(&options)
	if err != nil {
		return
	}
	err = r.Mode.AddOptions(&options, settings)
	if err != nil {
		return
	}
	err = r.Rules.AddOptions(&options)
	if err != nil {
		return
	}
	err = r.Scope.AddOptions(&options, r.Mode)
	if err != nil {
		return
	}
	err = settings.ProxySettings()
	if err != nil {
		return
	}
	err = settings.Write()
	if err != nil {
		return
	}
	f, err := addon.File.Post(settings.path())
	if err != nil {
		return
	}
	addon.Attach(f)
	return
}
