package main

import (
	"errors"
	"github.com/konveyor/analyzer-lsp/provider"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

//
// Settings - provider settings file.
type Settings []provider.Config

//
// Read file.
func (r *Settings) Read() (err error) {
	f, err := os.Open(r.path())
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := io.ReadAll(f)
	err = yaml.Unmarshal(b, r)
	return
}

//
// Write file.
func (r *Settings) Write() (err error) {
	f, err := os.Create(r.path())
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	b, err := yaml.Marshal(r)
	if err != nil {
		return
	}
	_, err = f.Write(b)
	return
}

//
// Location update the location on each provider.
func (r *Settings) Location(path string) {
	for i := range *r {
		p := &(*r)[i]
		p.InitConfig[0].Location = path
	}
}

//
// Mode update the mode on each provider.
func (r *Settings) Mode(mode provider.AnalysisMode) {
	for i := range *r {
		p := &(*r)[i]
		switch p.Name {
		case "java":
			p.InitConfig[0].AnalysisMode = mode
		}
	}
}

//
// MavenSettings set maven settings path.
func (r *Settings) MavenSettings(path string) {
	if path == "" {
		return
	}
	for i := range *r {
		p := &(*r)[i]
		switch p.Name {
		case "java":
			p.InitConfig[0].ProviderSpecificConfig["mavenSettingsFile"] = path
		}
	}
}

//
// ProxySettings set proxy settings.
func (r *Settings) ProxySettings() (err error) {
	var http, https string
	var excluded, noproxy []string
	http, excluded, err = r.getProxy("http")
	if err != nil {
		if errors.Is(err, &hub.NotFound{}) {
			noproxy = append(noproxy, excluded...)
		} else {
			return
		}
	}
	https, excluded, err = r.getProxy("https")
	if err != nil {
		if errors.Is(err, &hub.NotFound{}) {
			noproxy = append(noproxy, excluded...)
		} else {
			return
		}
	}
	for i := range *r {
		p := &(*r)[i]
		switch p.Name {
		case "java":
			d := p.InitConfig[0].ProviderSpecificConfig
			d["httpproxy"] = http
			d["httpsproxy"] = https
			d["noproxy"] = strings.Join(noproxy, ",")
		}
	}
	return
}

//
// getProxy set proxy settings.
func (r *Settings) getProxy(kind string) (url string, excluded []string, err error) {
	var p *api.Proxy
	var id *api.Identity
	p, err = addon.Proxy.Find(kind)
	if err != nil {
		return
	}
	if p.Identity != nil {
		id, err = addon.Identity.Get(p.Identity.ID)
		if err == nil {
			p.Host = id.User + ":" + id.Password + "@" + p.Host
		} else {
			return
		}
		excluded = append(
			excluded,
			p.Excluded...)
	}
	url = "http://" + p.Host
	if p.Port > 0 {
		url = url + ":" + strconv.Itoa(p.Port)
	}
	return
}

//
// Report self as activity.
func (r *Settings) Report() {
	b, _ := yaml.Marshal(r)
	addon.Activity("Settings: %s\n%s", r.path(), string(b))
}

//
// Path
func (r *Settings) path() (p string) {
	return path.Join(OptDir, "settings.json")
}
