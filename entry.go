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
	Globs        []Entry // only populated from glob'd input
}

// NewEntry takes a filepath and expands ~ as well as other relative paths to absolute and stats them returning Entry.
func NewEntry(inputPath string) (Entry, error) {

	inputPath = filepath.Clean(strings.TrimSpace(inputPath))

	// expand ~ paths
	if strings.HasPrefix(inputPath, "~") {
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
	} else {
		filenames, err := unglobInput(abs)
		if err != nil {
			return Entry{}, fmt.Errorf("error unglobbing input: %s, error: %w", abs, err)
		}

		var globs = make([]Entry, len(filenames))
		for i, file := range filenames {
			globs[i], err = NewEntry(file)
			if err != nil {
				return Entry{}, fmt.Errorf("error creating unglobbed entries, given input: %s, current file: %s, error: %w", abs, file, err)
			}
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

// OnlyNames returns a slice of absolute paths (strings) from a given Entry slice.
func OnlyNames(input []Entry) []string {
	var result = make([]string, len(input))
	for i, entry := range input {
		result[i] = entry.String()
	}
	return result
}

// unglobInput expands ~, and un-globs input.
func unglobInput(inputPath string) ([]string, error) {

	inputPath = filepath.Clean(strings.TrimSpace(inputPath))

	// expand ~ paths
	if strings.HasPrefix(inputPath, "~") {
		user, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("error getting current user, error: %s", err.Error())
		}
		inputPath = filepath.Join(user.HomeDir, strings.ReplaceAll(inputPath, "~", ""))
	}

	// try un-globing the input
	return filepath.Glob(inputPath)
}
