package main

import (
	"fmt"
	"os"
	"path"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/analyzer-lsp/core"
	"github.com/konveyor/tackle2-addon-analyzer/builder"
	addonprogress "github.com/konveyor/tackle2-addon-analyzer/progress"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gopkg.in/yaml.v2"
)

// Analyzer application analyzer.
type Analyzer struct {
	*Data
}

// Run analyzer.
func (r *Analyzer) Run() (insights *builder.Insights, deps *builder.Deps, err error) {

	analyzerOpts, err := r.options()
	if err != nil {
		return
	}
	log := logr.New("analyzer", r.Verbosity+4)
	analyzerOpts = append(analyzerOpts, core.WithLogger(log))

	analyzerOpts = append(analyzerOpts, core.WithReporters(addonprogress.NewAddonReporter(addon)))
	// If there are any errors from opts they will be set here.
	analyzer, err := core.NewAnalyzer(analyzerOpts...)
	if err != nil {
		addon.Error(api.TaskError{
			Severity:    "Error",
			Description: fmt.Sprintf("Unable to start Analyzer errs: %s", err.Error()),
		})
		return
	}
	defer analyzer.Stop()

	_, err = analyzer.ParseRules()
	if err != nil {
		return
	}

	err = analyzer.ProviderStart()
	if err != nil {
		return
	}

	for _, p := range analyzer.GetProviders() {
		log.Info("capabilities", "caps", p.Capabilities())
	}

	depOutput := path.Join(Dir, "deps.yaml")
	output := path.Join(Dir, "insights.yaml")

	results := analyzer.Run()
	if !r.Data.Mode.Discovery {
		depErr := analyzer.GetDependencies(depOutput, false)
		if depErr != nil {
			err = depErr
			return
		}
	}

	if r.Verbosity > 0 {
		// Create the files and post
		i, mErr := yaml.Marshal(results)
		if mErr != nil {
			err = mErr
			return
		}
		file, cErr := os.Create(output)
		if cErr != nil {
			err = cErr
			return
		}
		_, wErr := file.Write(i)
		file.Close()
		if wErr != nil {
			err = wErr
			return
		}

		f, pErr := addon.File.Post(output)
		if pErr != nil {
			err = pErr
			return
		}
		addon.Attach(f)

		if _, statErr := os.Stat(depOutput); statErr == nil {
			f, pErr = addon.File.Post(depOutput)
			if pErr != nil {
				err = pErr
				return
			}
			addon.Attach(f)
		}
	}
	insights, err = builder.NewInsights(results)
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
func (r *Analyzer) options() (options []core.AnalyzerOption, err error) {

	options = append(options, r.Mode.ToOption())
	options = append(options, r.Rules.ToOptions()...)
	options = append(options, r.Scope.ToOptions(r.Mode)...)
	settings := Settings{}
	err = settings.AppendExtensions(&r.Mode)
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
	f, pErr := addon.File.Post(settings.path())
	if pErr != nil {
		err = pErr
		return
	}
	addon.Attach(f)
	options = append(options, core.WithProviderConfigs(settings.Configs))
	return
}
