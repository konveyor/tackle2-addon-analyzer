package main

import (
	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
	"os"
	"path"
	"strings"
)

//
// Mode settings.
type Mode struct {
	Binary     bool   `json:"binary"`
	Artifact   string `json:"artifact"`
	WithDeps   bool   `json:"withDeps"`
	Repository repository.SCM
	//
	path struct {
		appDir string
		binary string
		maven  struct {
			settings string
		}
	}
}

//
// Build assets.
func (r *Mode) Build(application *api.Application) (err error) {
	if !r.Binary {
		err = r.fetchRepository(application)
		if err != nil {
			return
		}
		err = r.mavenSettings(application)
		if err != nil {
			return
		}
	} else {
		if r.Artifact != "" {
			err = r.getArtifact()
			return
		}
		if application.Binary != "" {
			err = r.mavenArtifact(application)
			return
		}
	}
	return
}

//
// AddOptions adds analyzer options.
func (r *Mode) AddOptions(settings *Settings) (err error) {
	if r.WithDeps {
		settings.MavenSettings(r.path.maven.settings)
		settings.Mode(provider.FullAnalysisMode)
	} else {
		settings.Mode(provider.SourceOnlyAnalysisMode)
	}
	if r.Binary {
		settings.Location(r.path.binary)
	} else {
		settings.Location(r.path.appDir)
	}
	return
}

//
// fetchRepository get SCM repository.
func (r *Mode) fetchRepository(application *api.Application) (err error) {
	if application.Repository == nil {
		err = &SoftError{Reason: "Application repository not defined."}
		return
	}
	SourceDir = path.Join(
		SourceDir,
		strings.Split(
			path.Base(
				application.Repository.URL),
			".")[0])
	r.path.appDir = path.Join(SourceDir, application.Repository.Path)
	var rp repository.SCM
	rp, nErr := repository.New(
		SourceDir,
		application.Repository,
		application.Identities)
	if nErr != nil {
		err = nErr
		return
	}
	err = rp.Fetch()
	return
}

//
// getArtifact get uploaded artifact.
func (r *Mode) getArtifact() (err error) {
	bucket := addon.Bucket()
	err = bucket.Get(r.Artifact, BinDir)
	r.path.binary = path.Join(BinDir, path.Base(r.Artifact))
	return
}

//
// mavenArtifact get maven artifact.
func (r *Mode) mavenArtifact(application *api.Application) (err error) {
	binDir := path.Join(BinDir, "maven")
	maven := repository.Maven{
		M2Dir:  M2Dir,
		BinDir: binDir,
		Remote: repository.Remote{
			Identities: application.Identities,
		},
	}
	err = maven.FetchArtifact(application.Binary)
	if err != nil {
		return
	}
	dir, nErr := os.ReadDir(binDir)
	if nErr != nil {
		err = nErr
		return
	}
	if len(dir) > 0 {
		r.path.binary = path.Join(binDir, dir[0].Name())
	}
	return
}

//
// mavenSettings writes maven settings.
func (r *Mode) mavenSettings(application *api.Application) (err error) {
	maven := repository.Maven{
		Remote: repository.Remote{
			Identities: application.Identities,
		},
	}
	settings, err := maven.Settings()
	if settings == "" || err != nil {
		return
	}
	p := path.Join(
		OptDir,
		"maven",
		"settings.xml")
	err = nas.MkDir(path.Dir(p), 0755)
	if err != nil {
		return
	}
	f, err := os.Create(p)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.WriteString(settings)
	if err != nil {
		return
	}
	r.path.maven.settings = p
	return
}
