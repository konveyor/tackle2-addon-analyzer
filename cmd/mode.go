package main

import (
	"errors"
	"path"
	"strings"

	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/konveyor/tackle2-hub/addon/command"
	"github.com/konveyor/tackle2-hub/addon/scm"
	"github.com/konveyor/tackle2-hub/api"
)

// Mode settings.
type Mode struct {
	Discovery  bool   `json:"discovery"`
	Binary     bool   `json:"binary"`
	Artifact   string `json:"artifact"`
	WithDeps   bool   `json:"withDeps"`
	Repository scm.SCM
	//
	path struct {
		appDir string
		binary string
	}
}

// Build assets.
func (r *Mode) Build(application *api.Application) (err error) {
	if !r.Binary {
		err = r.fetchRepository(application)
		return
	}
	if r.Artifact != "" {
		err = r.getArtifact()
		return
	}
	if application.Binary != "" {
		r.path.binary = application.Binary + "@" + BinDir
	}
	return
}

// AddOptions adds analyzer options.
func (r *Mode) AddOptions(options *command.Options, settings *Settings) (err error) {
	if r.WithDeps {
		settings.Mode(provider.FullAnalysisMode)
	} else {
		settings.Mode(provider.SourceOnlyAnalysisMode)
		options.Add("--no-dependency-rules")
	}
	return
}

// Location returns the location to be analyzed.
func (r *Mode) Location() (path string) {
	if r.Binary {
		path = r.path.binary
	} else {
		path = r.path.appDir
	}
	return
}

// fetchRepository get SCM repository.
func (r *Mode) fetchRepository(application *api.Application) (err error) {
	if application.Repository == nil {
		err = errors.New("Application repository not defined.")
		return
	}
	identity, _, err :=
		addon.Application.Identity(application.ID).Search().
			Direct("source").
			Indirect("source").
			Find()
	if err != nil {
		return
	}
	SourceDir = path.Join(
		SourceDir,
		strings.Split(
			path.Base(
				application.Repository.URL),
			".")[0])
	r.path.appDir = path.Join(SourceDir, application.Repository.Path)
	r.Repository, err = addon.SCM.New(
		SourceDir,
		*application.Repository,
		identity)
	if err != nil {
		return
	}
	err = r.Repository.Fetch()
	return
}

// getArtifact get uploaded artifact.
func (r *Mode) getArtifact() (err error) {
	bucket := addon.Bucket()
	err = bucket.Get(r.Artifact, BinDir)
	r.path.binary = path.Join(BinDir, path.Base(r.Artifact))
	return
}
