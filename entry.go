package path

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Entry is the currency of this package.
type Entry struct {
	FileInfo     fs.FileInfo
	AbsolutePath string
	Children     []Entry
}

// NewEntry is the public constructor for creating Entry. The levelsDeep param controls the level of recursion
// when collecting file info in subdirectories. levelsDeep == 0 will only create an entry for inputPath.
// Consider the number of files that may be under the root directory and the memory required to represent them
// when choosing this value.
func NewEntry(inputPath string, levelsDeep int, filters ...ListFilter) (Entry, error) {

	var root, err = newEntry(inputPath)
	if err != nil {
		return Entry{}, err
	}

	var currLevel = &root
	if currLevel.IsDir() && levelsDeep > 0 && len(root.Children) == 0 {

		if err := currLevel.populateChildren(levelsDeep); err != nil {
			return Entry{}, err
		}
	}

	return root, nil
}

// populateChildren recursively populates the children of a directory.
func (e *Entry) populateChildren(levels int, filters ...ListFilter) error {

	files, err := os.ReadDir(e.AbsolutePath)
	if err != nil {
		return err
	}

	var children = make([]Entry, len(files))
FileLoop:
	for i, file := range files {

		var entry, err = newEntry(filepath.Join(e.AbsolutePath, file.Name()))
		if err != nil {
			return err
		}

		// filter out the children we dont need ... sorry kids :(
		for _, fn := range filters {

			var accepted, err = fn.filter(entry)
			if err != nil {
				return fmt.Errorf("error filtering children: %w", err)
			}
			if !accepted {
				continue FileLoop
			}
		}

		children[i] = entry
	}

	e.Children = children

	if levels > 0 {
		for i, child := range e.Children {
			if child.IsDir() {
				err = e.Children[i].populateChildren(levels - 1)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// newEntry takes a filepath and expands ~ as well as other relative paths to absolute and stats them returning Entry.
func newEntry(inputPath string) (Entry, error) {

	var entry = Entry{}
	var err error
	inputPath = filepath.Clean(strings.TrimSpace(inputPath))

	inputPath, unglobbedFilenames, err := unglobInput(inputPath)
	if err != nil {
		return Entry{}, fmt.Errorf("error unglobbing input: %s, error: %w", entry.AbsolutePath, err)
	}

	entry.AbsolutePath, err = filepath.Abs(inputPath)
	if err != nil {
		return Entry{}, fmt.Errorf("error getting absolute path, error: %s", err.Error())
	}

	if len(unglobbedFilenames) <= 1 {

		entry.FileInfo, err = os.Stat(entry.AbsolutePath)
		if err != nil {
			return Entry{}, fmt.Errorf("error stating file: %s, error: %w", entry.AbsolutePath, err)
		}
	} else {

		entry.Children = make([]Entry, len(unglobbedFilenames))
		for i, file := range unglobbedFilenames {

			entry.Children[i], err = NewEntry(file, 0)
			if err != nil {
				return Entry{}, fmt.Errorf("error creating unglobbed entries, given input: %s, current file: %s, error: %w", entry.AbsolutePath, file, err)
			}
		}

		entry.FileInfo, err = os.Stat(filepath.Dir(entry.AbsolutePath)) // we use Dir() here because its globbed and will not work otherwise
		if err != nil {
			return Entry{}, fmt.Errorf("error stating file: %s, error: %w", entry.AbsolutePath, err)
		}
	}

	return entry, nil
}

// String fulfils the flag.Value interface https://pkg.go.dev/flag#Value.
func (e *Entry) String() string {
	return e.AbsolutePath
}

func (e *Entry) IsDir() bool {
	return e.FileInfo.IsDir()
}

// List recursively lists all files with optional filters. The root directory "inputPath" is excluded from the results.
func (e *Entry) Flatten() ([]Entry, error) {
	return collectChildern(*e)
}

func collectChildern(entry Entry) ([]Entry, error) {

	var entries = entry.Children

	if len(entry.Children) > 0 {

		for _, child := range entry.Children {

			var newChildren, err = collectChildern(child)
			if err != nil {
				return nil, err
			}

			entries = append(entries, newChildren...)
		}
	}

	return entries, nil
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
func unglobInput(inputPath string) (string, []string, error) {

	// expand ~ paths
	if strings.HasPrefix(inputPath, "~") {
		user, err := user.Current()
		if err != nil {
			return "", nil, fmt.Errorf("error getting current user, error: %w", err)
		}
		inputPath = filepath.Join(user.HomeDir, strings.ReplaceAll(inputPath, "~", ""))
	}

	// try un-globing the input
	globs, err := filepath.Glob(inputPath)
	return inputPath, globs, err
}
