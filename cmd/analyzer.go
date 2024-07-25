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
func (r *Analyzer) Run() (b *builder.Issues, err error) {
	output := path.Join(Dir, "report.yaml")
	cmd := command.New("/usr/local/bin/konveyor-analyzer")
	cmd.Options, err = r.options(output)
	if err != nil {
		return
	}
	if Verbosity > 0 {
		cmd.Reporter.Verbosity = command.LiveOutput
	}
	b = &builder.Issues{Path: output}
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
	}
	return
}

// options builds Analyzer options.
func (r *Analyzer) options(output string) (options command.Options, err error) {
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

// DepAnalyzer application analyzer.
type DepAnalyzer struct {
	*Data
}

// Run analyzer.
func (r *DepAnalyzer) Run() (b *builder.Deps, err error) {
	output := path.Join(Dir, "deps.yaml")
	cmd := command.New("/usr/local/bin/konveyor-analyzer-dep")
	cmd.Options, err = r.options(output)
	if err != nil {
		return
	}
	if Verbosity > 0 {
		cmd.Reporter.Verbosity = command.LiveOutput
	}
	b = &builder.Deps{Path: output}
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}

// options builds Analyzer options.
func (r *DepAnalyzer) options(output string) (options command.Options, err error) {
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
	}
	err = r.Mode.AddDepOptions(&options, settings)
	if err != nil {
		return
	}
	err = settings.Write()
	if err != nil {
		return
	}
	return
}
