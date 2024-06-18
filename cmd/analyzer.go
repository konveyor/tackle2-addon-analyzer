package main

import (
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
func (r *Analyzer) Run() (ir *builder.Issues, dr *builder.Deps, err error) {
	output := path.Join(Dir, "report.yaml")
	depOutput := path.Join(Dir, "deps.yaml")
	cmd := command.New("/usr/local/bin/konveyor-analyzer")
	cmd.Options, err = r.options(output, depOutput)
	if err != nil {
		return
	}
	if Verbosity > 0 {
		cmd.Reporter.Verbosity = command.LiveOutput
	}
	ir = &builder.Issues{Path: output}
	dr = &builder.Deps{Path: output}
	err = cmd.Run()
	return
}

// options builds Analyzer options.
func (r *Analyzer) options(output, depOutput string) (options command.Options, err error) {
	settings := &Settings{}
	err = settings.Read()
	if err != nil {
		return
	}
	options = command.Options{
		"--provider-settings",
		settings.path(),
		"--output-file",
		output,
		"—dep-output-file",
		depOutput,
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
	err = r.Scope.AddOptions(&options)
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
