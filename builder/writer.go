package builder

import (
	"io"

	"gopkg.in/yaml.v2"
)

// Writer provides safe output writer.
type Writer struct {
	wrapped io.Writer
	errors  []error
}

// write the string.
func (w *Writer) Write(s string) {
	_, err := io.WriteString(w.wrapped, s)
	if err != nil {
		w.addError(err)
		return
	}
}

// Encode and write object.
func (w *Writer) Encode(object any) {
	_, err := w.wrapped.Write([]byte("---\n"))
	if err != nil {
		w.addError(err)
		return
	}
	encoder := yaml.NewEncoder(w.wrapped)
	err = encoder.Encode(object)
	if err != nil {
		w.addError(err)
		return
	}
	err = encoder.Close()
	if err != nil {
		w.addError(err)
		return
	}
	return
}

// Error returns the first error.
func (w *Writer) Error() (err error) {
	if len(w.errors) > 0 {
		err = w.errors[0]
	}
	return
}

// addError adds errors to the list.
func (w *Writer) addError(err error) {
	if err != nil {
		w.errors = append(w.errors, err)
	}
}
