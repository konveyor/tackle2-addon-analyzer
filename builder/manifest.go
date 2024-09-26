package builder

import (
	"os"

	"github.com/konveyor/tackle2-hub/api"
	"gopkg.in/yaml.v2"
)

// Manifest file.
type Manifest struct {
	Analysis api.Analysis
	Issues   *Issues
	Deps     *Deps
	Path     string
}

// Write manifest file.
func (m *Manifest) Write() (err error) {
	m.Path = "manifest.yaml"
	file, err := os.Create(m.Path)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()
	_, _ = file.Write([]byte(api.BeginMainMarker))
	_, _ = file.Write([]byte{'\n'})
	encoder := yaml.NewEncoder(file)
	err = encoder.Encode(m.Analysis)
	if err != nil {
		return
	}
	_, _ = file.Write([]byte(api.EndMainMarker))
	_, _ = file.Write([]byte{'\n'})
	err = m.Issues.Write(file)
	if err != nil {
		return
	}
	err = m.Deps.Write(file)
	if err != nil {
		return
	}
	return
}
