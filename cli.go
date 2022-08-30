package path

import (
	"strings"
)

// Path is a type whos sole purpose is to furfil the flag interface
type Path struct {
	Input string
	Files []Entry
}

// String fulfils the flag.Value interface https://pkg.go.dev/flag#Value
func (p Path) String() string {
	var b = strings.Builder{}
	for i, f := range p.Files {
		if i != 0 {
			b.WriteString(" ")
		}
		b.WriteString(f.FileInfo.Name())
	}
	return b.String()
}

// Get fulfils the flag.Getter interface https://pkg.go.dev/flag#Getter
func (p *Path) Get() []Entry {
	return p.Files
}

// Set fulfils the flag.Value interface https://pkg.go.dev/flag#Value
func (p *Path) Set(s string) error {
	p.Input = s
	var files, err = ListFiles(s)
	p.Files = files
	return err
}
