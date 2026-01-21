package main

import (
	"os"
	"path"
	"time"

	"github.com/konveyor/tackle2-addon-analyzer/builder"
	hub "github.com/konveyor/tackle2-hub/shared/addon"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/env"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"gopkg.in/yaml.v2"
)

var (
	addon     = hub.Addon
	BinDir    = ""
	SharedDir = ""
	CacheDir  = ""
	SourceDir = ""
	Dir       = ""
	M2Dir     = ""
	RuleDir   = ""
	OptDir    = ""
	Source    = "Analysis"
	Verbosity = 0
)

func init() {
	Dir, _ = os.Getwd()
	OptDir = path.Join(Dir, "opt")
	SharedDir = env.Get(hub.EnvSharedDir, "/tmp/shared")
	CacheDir = env.Get(hub.EnvCacheDir, "/tmp/cache")
	SourceDir = path.Join(SharedDir, "source")
	RuleDir = path.Join(Dir, "rules")
	BinDir = path.Join(SharedDir, "bin")
	M2Dir = path.Join(CacheDir, "m2")
}

// Data Addon data passed in the secret.
type Data struct {
	// Verbosity level.
	Verbosity int `json:"verbosity"`
	// Profile id.
	Profile api.Ref `json:"profile"`
	// Mode options.
	Mode Mode `json:"mode"`
	// Scope options.
	Scope Scope `json:"scope"`
	// Rules options.
	Rules Rules `json:"rules"`
	// Tagger options.
	Tagger Tagger `json:"tagger"`
}

// main
func main() {
	addon.Run(func() (err error) {
		addon.Activity("OptDir:    %s", OptDir)
		addon.Activity("SharedDir: %s", SharedDir)
		addon.Activity("CacheDir:  %s", CacheDir)
		addon.Activity("SourceDir: %s", SourceDir)
		addon.Activity("RuleDir:   %s", RuleDir)
		addon.Activity("BinDir:    %s", BinDir)
		addon.Activity("M2Dir:     %s", M2Dir)
		//
		// Get the addon data associated with the task.
		d := &Data{}
		err = addon.DataWith(d)
		if err == nil {
			Verbosity = d.Verbosity
		} else {
			return
		}
		//
		// Create directories.
		for _, dir := range []string{BinDir, M2Dir, RuleDir, OptDir} {
			err = nas.MkDir(dir, 0755)
			if err != nil {
				return
			}
		}
		//
		// Fetch application.
		addon.Activity("Fetching application.")
		application, err := addon.Task.Application()
		if err != nil {
			return
		}
		//
		// Apply profile.
		err = applyProfile(d)
		if err != nil {
			return
		}
		//
		// Build assets.
		err = d.Mode.Build(application)
		if err != nil {
			return
		}
		err = d.Rules.Build()
		if err != nil {
			return
		}
		//
		// Run the analyzer.
		analyzer := Analyzer{}
		analyzer.Data = d
		insights, deps, err := analyzer.Run()
		if err != nil {
			return
		}
		//
		// RuleError
		ruleErr := insights.RuleError()
		ruleErr.Report()
		//
		// Update application.
		err = updateApplication(d, application.ID, insights, deps)
		if err != nil {
			return
		}

		addon.Activity("Done.")

		return
	})
}

// applyProfile fetch and apply profile when specified.
func applyProfile(d *Data) (err error) {
	if d.Profile.ID == 0 {
		return
	}
	d.Mode = Data{}.Mode
	d.Scope = Data{}.Scope
	d.Rules = Data{}.Rules
	p, err := addon.AnalysisProfile.Get(d.Profile.ID)
	if err != nil {
		return
	}
	b, _ := yaml.Marshal(p)
	addon.Activity(
		"Using profile (id=%d): %s\n%s",
		p.ID,
		p.Name,
		string(b))
	err = d.Mode.With(&p.Mode)
	if err != nil {
		return
	}
	err = d.Scope.With(&p.Scope)
	if err != nil {
		return
	}
	err = d.Rules.With(&p.Rules)
	if err != nil {
		return
	}
	b, _ = yaml.Marshal(d)
	addon.Activity(
		"Using configuration:\n%s",
		string(b))
	return
}

// updateApplication creates analysis report and updates
// the application facts and tags.
func updateApplication(d *Data, appId uint, insights *builder.Insights, deps *builder.Deps) (err error) {
	//
	// Tags.
	if d.Tagger.Enabled {
		if d.Tagger.Source == "" {
			d.Tagger.Source = Source
		}
		err = d.Tagger.Update(appId, insights.Tags())
		if err != nil {
			return
		}
	}
	if d.Mode.Discovery {
		return
	}
	//
	// Analysis.
	manifest := builder.Manifest{
		Analysis: api.Analysis{},
		Insights: insights,
		Deps:     deps,
	}
	if d.Mode.Repository != nil {
		manifest.Analysis.Commit, err = d.Mode.Repository.Head()
		if err != nil {
			return
		}
	}
	err = manifest.Write()
	if err != nil {
		return
	}
	mark := time.Now()
	reported, err := addon.Application.
		Select(appId).
		Analysis.
		Upload(manifest.Path, api.MIMEYAML)
	if err != nil {
		return
	}
	addon.Activity("Analysis %d reported. duration: %s", reported.ID, time.Since(mark))
	// Facts.
	err = addon.Application.Select(appId).
		Fact.
		Source(Source).
		Replace(insights.Facts())
	if err == nil {
		addon.Activity("Facts updated.")
	}
	return
}
