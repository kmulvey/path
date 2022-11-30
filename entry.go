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

// Entry is the currency of this package.
type Entry struct {
	FileInfo     fs.FileInfo
	AbsolutePath string
}

// NewEntry takes a filepath and expands ~ as well as other relative paths to absolute and stats them returning Entry.
func NewEntry(inputPath string) (Entry, error) {

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

func (e *Entry) String() string {
	return e.AbsolutePath
}

func (e *Entry) IsDir() bool {
	return e.FileInfo.IsDir()
}

func Contains(input []Entry, needle string) bool {
	for _, entry := range input {
		if entry.AbsolutePath == needle {
			return true
		}
	}
	return false
}

// OnlyDirs filters an input slice by returning a slice containing only directories.
func OnlyDirs(input []Entry) []Entry {
	var result []Entry
	for _, entry := range input {
		if entry.FileInfo.IsDir() {
			result = append(result, entry)
		}
	}
	return result
}

// OnlyFiles filters an input slice by returning a slice containing only files.
func OnlyFiles(input []Entry) []Entry {
	var result []Entry
	for _, entry := range input {
		if !entry.FileInfo.IsDir() {
			result = append(result, entry)
		}
	}
	return result
}

// OnlyNames returns a slice of absolute paths (strings) from a given Entry slice.
func OnlyNames(input []Entry) []string {
	var result = make([]string, len(input))
	for i, entry := range input {
		result[i] = entry.String()
	}
	return result
}
