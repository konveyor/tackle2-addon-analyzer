package main

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/konveyor/analyzer-lsp/provider"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"gopkg.in/yaml.v2"
)

// Settings - provider settings file.
type Settings []provider.Config

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

// Location update the location on each provider.
func (r *Settings) Location(path string) {
	for i := range *r {
		p := &(*r)[i]
		p.InitConfig[0].Location = path
	}
}

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

// ProxySettings set proxy settings.
func (r *Settings) ProxySettings() (err error) {
	var http, https string
	var excluded, noproxy []string
	http, excluded, err = r.getProxy("http")
	if err == nil {
		noproxy = append(
			noproxy,
			excluded...)
	} else {
		return
	}
	https, excluded, err = r.getProxy("https")
	if err == nil {
		noproxy = append(
			noproxy,
			excluded...)
	} else {
		return
	}
	for i := range *r {
		p := &(*r)[i]
		switch p.Name {
		case "java":
			d := p.InitConfig[0].ProviderSpecificConfig
			if http != "" {
				d["httpproxy"] = http
			}
			if https != "" {
				d["httpsproxy"] = https
			}
			if len(noproxy) > 0 {
				d["noproxy"] = strings.Join(noproxy, ",")
			}
		}
	}
	return
}

// getProxy set proxy settings.
func (r *Settings) getProxy(kind string) (url string, excluded []string, err error) {
	var p *api.Proxy
	var id *api.Identity
	var user, password string
	p, err = addon.Proxy.Find(kind)
	if err != nil {
		if errors.Is(err, &hub.NotFound{}) {
			err = nil
			return
		}
	}
	if p.Host == "" {
		return
	}
	if p.Identity != nil {
		id, err = addon.Identity.Get(p.Identity.ID)
		if err == nil {
			user = id.User
			password = id.Password
		} else {
			return
		}
	}
	host := p.Host
	excluded = p.Excluded
	if user != "" && password != "" {
		host = user + ":" + password + "@" + host
	}
	if p.Port > 0 {
		host += ":" + strconv.Itoa(p.Port)
	}
	url = kind + "://" + host
	return
}

// Report self as activity.
func (r *Settings) Report() {
	b, _ := yaml.Marshal(r)
	addon.Activity("Settings: %s\n%s", r.path(), string(b))
}

// Path
func (r *Settings) path() (p string) {
	return path.Join(OptDir, "settings.json")
}
