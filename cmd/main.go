package main

import (
	"errors"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
	"k8s.io/utils/env"
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
	SharedDir = env.GetString(hub.EnvSharedDir, "/tmp/shared")
	CacheDir = env.GetString(hub.EnvCacheDir, "/tmp/cache")
	SourceDir = path.Join(SharedDir, "source")
	RuleDir = path.Join(Dir, "rules")
	BinDir = path.Join(SharedDir, "bin")
	M2Dir = path.Join(CacheDir, "m2")
}

// Data Addon data passed in the secret.
type Data struct {
	// Verbosity level.
	Verbosity int `json:"verbosity"`
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
		// SSH
		agent := ssh.Agent{}
		err = agent.Start()
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
		// Run analysis.
		analyzer := Analyzer{}
		analyzer.Data = d
		issues, err := analyzer.Run()
		if err != nil {
			return
		}
		if !d.Mode.Discovery {
			depAnalyzer := DepAnalyzer{}
			depAnalyzer.Data = d
			deps, dErr := depAnalyzer.Run()
			if dErr != nil {
				err = dErr
				return
			}
			//
			// Post report.
			appAnalysis := addon.Application.Analysis(application.ID)
			mark := time.Now()
			analysis := &api.Analysis{}
			err = appAnalysis.Create(
				analysis,
				binding.MIMEYAML,
				issues.Reader(),
				deps.Reader())
			if err == nil {
				addon.Activity("Analysis reported. duration: %s", time.Since(mark))
			} else {
				ruleErr := &RuleError{}
				if errors.As(err, &ruleErr) {
					ruleErr.Report()
					err = nil
				}
				return
			}
			//
			// Facts
			facts := addon.Application.Facts(application.ID)
			facts.Source(Source)
			err = facts.Replace(issues.Facts())
			if err == nil {
				addon.Activity("Facts updated.")
			} else {
				return
			}
		}

		//
		// Tags.
		if d.Tagger.Enabled {
			if d.Tagger.Source == "" {
				d.Tagger.Source = Source
			}
			err = d.Tagger.Update(application.ID, issues.Tags())
			if err != nil {
				return
			}
		}

		addon.Activity("Done.")

		return
	})
}
