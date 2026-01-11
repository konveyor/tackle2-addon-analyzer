package main

import (
	"os"
	"path"

	"github.com/konveyor/tackle2-addon-analyzer/builder"
	"github.com/konveyor/tackle2-hub/addon/command"
	"github.com/konveyor/tackle2-hub/env"
)

var (
	AnalyzerBin = ""
)

func init() {
	AnalyzerBin = env.Get(
		"ANALYZER",
		"/usr/local/bin/konveyor-analyzer")
}

// Analyzer application analyzer.
type Analyzer struct {
	*Data
}

// Run analyzer.
func (r *Analyzer) Run() (insights *builder.Insights, deps *builder.Deps, err error) {
	output := path.Join(Dir, "insights.yaml")
	depOutput := path.Join(Dir, "deps.yaml")
	cmd := command.New(AnalyzerBin)
	cmd.Options, err = r.options(output, depOutput)
	if err != nil {
		return
	}
	if Verbosity > 0 {
		if w, cast := cmd.Writer.(*command.Writer); cast {
			w.Reporter().Verbosity = command.LiveOutput
		}
	}
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
	insights, err = builder.NewInsights(output)
	if err != nil {
		return
	}
	deps, err = builder.NewDeps(depOutput)
	if err != nil {
		return
	}
	return
}

// options builds Analyzer options.
func (r *Analyzer) options(output, depOutput string) (options command.Options, err error) {
	settings := &Settings{}
	err = settings.AppendExtensions(&r.Mode)
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
