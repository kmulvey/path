package path

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var MaxDepth uint8 = 255 // arbitrary, but hopefully enough

// Entry is the currency of this package.
type Entry struct {
	FileInfo     fs.FileInfo
	AbsolutePath string
	Children     []Entry
}

// NewEntry is the public constructor for creating an Entry. The levelsDeep param controls the level of recursion
// when collecting file info in subdirectories. levelsDeep == 0 will only create an entry for inputPath.
// Consider the number of files that may be under the root directory and the memory required to represent them
// when choosing this value.
func NewEntry(inputPath string, levelsDeep uint8, filters ...EntriesFilter) (Entry, error) {

	var root, err = newEntry(inputPath)
	if err != nil {
		return Entry{}, err
	}

	var currLevel = &root
	if currLevel.IsDir() && levelsDeep > 0 && len(root.Children) == 0 {

		if err := currLevel.populateChildren(levelsDeep, filters...); err != nil {
			return Entry{}, err
		}
	}

	return root, nil
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

		entry.FileInfo, err = os.Lstat(entry.AbsolutePath)
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

		entry.FileInfo, err = os.Lstat(filepath.Dir(entry.AbsolutePath)) // we use Dir() here because its globbed and will not work otherwise
		if err != nil {
			return Entry{}, fmt.Errorf("error stating file: %s, error: %w", entry.AbsolutePath, err)
		}
	}

	return entry, nil
}

// populateChildren recursively populates the children of an Entry.
func (e *Entry) populateChildren(levels uint8, filters ...EntriesFilter) error {

	files, err := os.ReadDir(e.AbsolutePath)
	if err != nil {
		return err
	}

	var children []Entry

FileLoop:
	for _, file := range files {

		var entry, err = newEntry(filepath.Join(e.AbsolutePath, file.Name()))
		if err != nil {
			return err
		}

		// we dont filter dirs or symlinks because this is a recursive func and we may miss files deeper in the dir structure
		if !file.IsDir() && entry.FileInfo.Mode()&os.ModeSymlink != fs.ModeSymlink {
			// filter out the children we dont need ... sorry kids :(
			for _, fn := range filters {
				var accepted = fn.filter(entry)
				if !accepted {
					continue FileLoop
				}
			}
		}

		children = append(children, entry)
	}

	e.Children = children
	levels--

	if levels > 0 {
		for i, child := range e.Children {
			if child.IsDir() || child.FileInfo.Mode()&os.ModeSymlink == fs.ModeSymlink {
				err = e.Children[i].populateChildren(levels-1, filters...)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// String fulfils the flag.Value interface https://pkg.go.dev/flag#Value.
func (e *Entry) String() string {
	return e.AbsolutePath
}

func (e *Entry) IsDir() bool {
	return e.FileInfo.IsDir()
}

// List recursively lists all files with optional filters. If includeRoot is true the root directory "inputPath" is included in the results.
func (e *Entry) Flatten(includeRoot bool) ([]Entry, error) {
	var arr, err = collectChildern(*e)
	if err != nil {
		return nil, err
	}

	if includeRoot {
		return append(arr, *e), nil
	}

	return arr, nil
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
