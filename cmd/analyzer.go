package main

import (
	"github.com/konveyor/tackle2-addon-analyzer/builder"
	"github.com/konveyor/tackle2-addon/command"
	"path"
)

type RuleError = builder.RuleError

//
// Analyzer application analyzer.
type Analyzer struct {
	*Data
}

//
// Run analyzer.
func (r *Analyzer) Run() (b *builder.Issues, err error) {
	output := path.Join(Dir, "report.yaml")
	cmd := command.Command{Path: "/usr/bin/konveyor-analyzer"}
	cmd.Options, err = r.options(output)
	if err != nil {
		return
	}
	b = &builder.Issues{Path: output}
	err = cmd.Run()
	return
}

//
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
	err = r.Mode.AddOptions(settings)
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
	settings.Report()
	return
}

//
// DepAnalyzer application analyzer.
type DepAnalyzer struct {
	*Data
}

//
// Run analyzer.
func (r *DepAnalyzer) Run() (b *builder.Deps, err error) {
	output := path.Join(Dir, "deps.yaml")
	cmd := command.Command{Path: "/usr/bin/konveyor-analyzer-dep"}
	cmd.Options, err = r.options(output)
	if err != nil {
		return
	}
	b = &builder.Deps{Path: output}
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}

//
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
	err = r.Mode.AddOptions(settings)
	if err != nil {
		return
	}
	err = settings.Write()
	if err != nil {
		return
	}
	return
}
