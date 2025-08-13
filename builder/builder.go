package builder

import (
	"io"

	"gopkg.in/yaml.v2"
)

type Builder struct {
	errors []error
}

// write the string.
func (b *Builder) write(writer io.Writer, s string) {
	_, err := io.WriteString(writer, s)
	if err != nil {
		b.errors = append(b.errors, err)
		return
	}
}

// encode object.
func (b *Builder) encode(writer io.Writer, r any) {
	_, err := writer.Write([]byte("---\n"))
	if err != nil {
		b.addError(err)
		return
	}
	encoder := yaml.NewEncoder(writer)
	err = encoder.Encode(r)
	if err != nil {
		b.addError(err)
		return
	}
	err = encoder.Close()
	if err != nil {
		b.addError(err)
		return
	}
	return
}

// addError adds errors to the list.
func (b *Builder) addError(err error) {
	if err != nil {
		b.errors = append(b.errors, err)
	}
}

// error returns the first error.
func (b *Builder) error() (err error) {
	if len(b.errors) > 0 {
		err = b.errors[0]
	}
	return
}
