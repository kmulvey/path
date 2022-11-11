package path

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

// Path is a type whos sole purpose is to furfil the flag interface
type Path struct {
	GivenInput   string // exactly what the user typed
	ComputedPath Entry  // converts relative paths to absolute
	Files        []Entry
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
	p.GivenInput = s

	s = filepath.Clean(s)

	var files, err = ListFiles(s)
	if err != nil {
		return err
	}
	p.Files = files

	p.ComputedPath, err = inputToEntry(s)
	return err
}

// inputToEntry expands ~ and other relative paths to absolute
func inputToEntry(inputPath string) (Entry, error) {

	// expand ~ paths
	if strings.Contains(inputPath, "~") {
		user, err := user.Current()
		if err != nil {
			return Entry{}, fmt.Errorf("error getting current user, error: %s", err.Error())
		}
		inputPath = filepath.Join(user.HomeDir, strings.ReplaceAll(inputPath, "~", ""))
	}

	var abs, err = filepath.Abs(inputPath)
	if err != nil {
		return Entry{}, fmt.Errorf("error getting absolute path, error: %s", err.Error())
	}

	var stat fs.FileInfo
	var hasGlob = regexp.MustCompile(`\*|\!|\?|\[|\]`)
	if !hasGlob.MatchString(abs) {
		stat, err = os.Stat(abs)
		if err != nil {
			return Entry{}, fmt.Errorf("error stating file: %s, error: %w", abs, err)
		}
	}

	return Entry{AbsolutePath: abs, FileInfo: stat}, nil
}
