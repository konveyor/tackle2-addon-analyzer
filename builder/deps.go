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

// Reader returns a reader.
func (b *Deps) Reader() (r io.Reader) {
	r, w := io.Pipe()
	go func() {
		var err error
		defer func() {
			if err != nil {
				_ = w.CloseWithError(err)
			} else {
				_ = w.Close()
			}
		}()
		err = b.Write(w)
	}()
	return
}

// Write deps to the writer.
func (b *Deps) Write(writer io.Writer) (err error) {
	input, err := b.read()
	if err != nil {
		return
	}
	encoder := yaml.NewEncoder(writer)
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
	return
}

// read dependencies.
func (b *Deps) read() (input []output.DepsFlatItem, err error) {
	input = []output.DepsFlatItem{}
	f, err := os.Open(b.Path)
	if err != nil {
		if os.IsNotExist(err) {
			addon.Activity(err.Error())
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
