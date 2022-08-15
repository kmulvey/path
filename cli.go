package path

import (
	"io/fs"
	"strings"
)

type Path struct {
	Input string
	Files []fs.DirEntry
}

// String fulfils the flag.Value interface https://pkg.go.dev/flag#Value
func (p Path) String() string {
	var b = strings.Builder{}
	for _, f := range p.Files {
		b.WriteString(f.Name())
		b.WriteString(" ")
	}
	return b.String()
}

// Get fulfils the flag.Getter interface https://pkg.go.dev/flag#Getter
func (p *Path) Get() []fs.DirEntry {
	return p.Files
}

// Set fulfils the flag.Value interface https://pkg.go.dev/flag#Value
func (p *Path) Set(s string) error {
	var files, err = ListFiles(s)
	p.Files = files
	return err
}
