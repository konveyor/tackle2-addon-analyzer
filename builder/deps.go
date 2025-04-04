package builder

import (
	"io"
	"os"

	output "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	"github.com/konveyor/tackle2-hub/api"
	"gopkg.in/yaml.v2"
)

// NewDeps returns a new deps builder.
func NewDeps(path string) (b *Deps, err error) {
	b = &Deps{}
	err = b.read(path)
	return
}

// Deps builds dependencies.
type Deps struct {
	input []output.DepsFlatItem
}

// Write deps section.
func (b *Deps) Write(writer io.Writer) (err error) {
	encoder := yaml.NewEncoder(writer)
	_, _ = writer.Write([]byte(api.BeginDepsMarker))
	_, _ = writer.Write([]byte{'\n'})
	for _, p := range b.input {
		for _, d := range p.Dependencies {
			err = encoder.Encode(
				&api.TechDependency{
					Provider: p.Provider,
					Indirect: d.Indirect,
					Name:     d.Name,
					Version:  d.Version,
					SHA:      d.ResolvedIdentifier,
					Labels:   d.Labels,
				})
			if err != nil {
				return
			}
		}
	}
	_, _ = writer.Write([]byte(api.EndDepsMarker))
	_, _ = writer.Write([]byte{'\n'})
	return
}

// read dependencies.
func (b *Deps) read(path string) (err error) {
	b.input = []output.DepsFlatItem{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			addon.Log.Info(err.Error())
			err = nil
		}
		return
	}
	defer func() {
		_ = f.Close()
	}()
	d := yaml.NewDecoder(f)
	err = d.Decode(&b.input)
	return
}
