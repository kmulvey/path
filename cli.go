package path

import (
	"path/filepath"
	"strings"
)

// Path is a type whose sole purpose is to furfil the flag interface.
type Path struct {
	GivenInput   string // exactly what the user typed
	ComputedPath Entry  // converts relative paths to absolute
	Files        []Entry
}

// String fulfils the flag.Value interface https://pkg.go.dev/flag#Value.
func (p Path) String() string {
	var b = strings.Builder{}
	for i, f := range p.Files {
		if i != 0 {
			b.WriteString(" ")
		}
		b.WriteString(f.String())
	}
	return b.String()
}

// Get fulfils the flag.Getter interface https://pkg.go.dev/flag#Getter.
func (p *Path) Get() []Entry {
	return p.Files
}

// Set fulfils the flag.Value interface https://pkg.go.dev/flag#Value.
func (p *Path) Set(s string) error {
	p.GivenInput = s

	s = filepath.Clean(s)

	var files, err = ListFiles(s)
	if err != nil {
		return err
	}
	p.Files = files

	p.ComputedPath, err = NewEntry(s)
	return err
}
