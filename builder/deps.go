package builder

import (
	"io"
	"os"

	output "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	"github.com/konveyor/tackle2-hub/api"
	"gopkg.in/yaml.v2"
)

// Deps builds dependencies.
type Deps struct {
	Path string
}

// Write deps section.
func (b *Deps) Write(writer io.Writer) (err error) {
	input, err := b.read()
	if err != nil {
		return
	}
	encoder := yaml.NewEncoder(writer)
	_, _ = writer.Write([]byte(api.BeginDepsMarker))
	_, _ = writer.Write([]byte{'\n'})
	for _, p := range input {
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
func (b *Deps) read() (input []output.DepsFlatItem, err error) {
	input = []output.DepsFlatItem{}
	f, err := os.Open(b.Path)
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
	bfr, err := io.ReadAll(f)
	err = yaml.Unmarshal(bfr, &input)
	return
}
