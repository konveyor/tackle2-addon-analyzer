package main

import (
	"github.com/konveyor/analyzer-lsp/provider"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

const (
	SETTINGS = "/addon/opt/settings.json"
)

//
// Settings - provider settings file.
type Settings []provider.Config

//
// Read file.
func (r *Settings) Read(path string) (err error) {
	f, err := os.Open(path)
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
func (r *Settings) Write(path string) (err error) {
	f, err := os.Create(path)
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
