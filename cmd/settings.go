package main

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/konveyor/analyzer-lsp/provider"
	hub "github.com/konveyor/tackle2-hub/shared/addon"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gopkg.in/yaml.v2"
)

const (
	// Builtin namespace.
	Builtin = "builtin"
	// BuiltinLocation The (code) Location passed to the provider.
	BuiltinLocation = Builtin + ".location"
)

// Settings - provider settings file.
type Settings struct {
	index   int
	content []provider.Config
}

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
	err = yaml.Unmarshal(b, &r.content)
	if err != nil {
		return
	}
	r.index = len(r.content)
	return
}

// AppendExtensions adds extension fragments.
func (r *Settings) AppendExtensions(mode *Mode) (err error) {
	addon, err := addon.Addon(true)
	if err != nil {
		return
	}
	for _, extension := range addon.Extensions {
		var md *Metadata
		md, err = r.metadata(&extension)
		if r.hasProvider(&md.Provider) {
			continue
		}
		builtin := r.injectBuiltins(md, mode)
		injector := ResourceInjector{}
		injector.Use(builtin)
		err = injector.Inject(md)
		if err != nil {
			return
		}
		r.content = append(r.content, md.Provider)
	}
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
	b, err := yaml.Marshal(r.content)
	if err != nil {
		return
	}
	_, err = f.Write(b)
	return
}

// Mode update the mode on each provider.
func (r *Settings) Mode(mode provider.AnalysisMode) {
	extensions := r.content[r.index:]
	for i := range extensions {
		p := extensions[i]
		for i := range p.InitConfig {
			init := &p.InitConfig[i]
			if init.AnalysisMode == "" {
				init.AnalysisMode = mode
			}
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
	if len(http)+len(https) == 0 {
		return
	}
	extensions := r.content[r.index:]
	for i := range extensions {
		p := &extensions[i]
		p.Proxy = &provider.Proxy{
			HTTPProxy:  http,
			HTTPSProxy: https,
			NoProxy: strings.Join(
				noproxy,
				","),
		}
	}
	return
}

// injectBuiltins injects `builtin` field values.
func (r *Settings) injectBuiltins(md *Metadata, mode *Mode) (builtin map[string]any) {
	builtin = make(map[string]any)
	list := md.Provider.InitConfig
	for i := range list {
		in := &list[i]
		in.Location = mode.Location()
		builtin[BuiltinLocation] = in.Location
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

// Path returns the file path.
func (r *Settings) path() (p string) {
	return path.Join(OptDir, "settings.yaml")
}

// hasProvider returns true when the provider found.
func (r *Settings) hasProvider(p *provider.Config) (found bool) {
	for i := range r.content {
		if r.content[i].Name == p.Name {
			found = true
			break
		}
	}
	return
}

// metadata returns the metadata object within the extension.
func (r *Settings) metadata(extension *api.Extension) (md *Metadata, err error) {
	injector := Injector{}
	mp := injector.asMap(extension.Metadata)
	md = &Metadata{}
	err = injector.object(mp, md)
	if err != nil {
		return
	}
	return
}
