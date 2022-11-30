package path

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// preProcessInput expands ~, and un-globs input.
func preProcessInput(inputPath string) ([]string, error) {

	// expand ~ paths
	if strings.Contains(inputPath, "~") {
		user, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("error getting current user, error: %s", err.Error())
		}
		inputPath = filepath.Join(user.HomeDir, strings.ReplaceAll(inputPath, "~", ""))
	}

	// try un-globing the input
	return filepath.Glob(inputPath)
}

// ListFiles recursively lists all files with optional filters. The root directory "inputPath" is excluded from the results.
func ListFiles(inputPath string, filters ...FilesFilter) ([]Entry, error) {
	var allFiles []Entry

	var globFiles, err = preProcessInput(inputPath)
	if err != nil {
		return nil, fmt.Errorf("Error from pre-processing: %w", err)
	}

	for _, gf := range globFiles {
		err = filepath.Walk(gf, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Walk error in dir: %s, error: %w", path, err)
			}
			// do not include the root dir
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("Walk error in dir stating file: %s, error: %w", path, err)
			}
			if gf == path && stat.IsDir() {
				return nil
			}

			var entry = Entry{AbsolutePath: path, FileInfo: info}

			// try all the filter funcs
			for _, fn := range filters {
				var accepted, err = fn.filter(entry)
				if err != nil {
					return err
				}
				if !accepted {
					return nil
				}
			}

			allFiles = append(allFiles, entry)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}

//////////////////////////////////////////////////////////////////

// FilesFilter interface facilitates filtering of file events.
type FilesFilter interface {
	filter(Entry) (bool, error)
}

// TrueFilesFilter always returns true, helpful for tests.
type TrueFilesFilter struct{}

func (tf TrueFilesFilter) filter(entry Entry) (bool, error) {
	return true, nil
}

// FalseFilesFilter always returns false, helpful for tests.
type FalseFilesFilter struct{}

func (ff FalseFilesFilter) filter(entry Entry) (bool, error) {
	return false, nil
}

// RegexFilesFilter filters Entry by matching file names to a given regex.
type RegexFilesFilter struct {
	regex *regexp.Regexp
}

func NewRegexFilesFilter(filterRegex *regexp.Regexp) RegexFilesFilter {
	return RegexFilesFilter{regex: filterRegex}
}

func (rf RegexFilesFilter) filter(entry Entry) (bool, error) {
	return rf.regex.MatchString(entry.String()), nil
}

// DateFilesFilter filters Entry by matching ensuring ModTime is within the given date range.
type DateFilesFilter struct {
	from time.Time
	to   time.Time
}

func NewDateFilesFilter(from, to time.Time) DateFilesFilter {
	return DateFilesFilter{from: from, to: to}
}

func (df DateFilesFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.ModTime().Before(df.from) || entry.FileInfo.ModTime().After(df.to) {
		return false, nil
	}
	return true, nil
}

// SkipMapFilesFilter filters Entry by ensuring the given file is NOT within the given map.
type SkipMapFilesFilter struct {
	skipMap map[string]struct{}
}

func NewSkipMapFilesFilter(skipMap map[string]struct{}) SkipMapFilesFilter {
	return SkipMapFilesFilter{skipMap: skipMap}
}

func (smf SkipMapFilesFilter) filter(entry Entry) (bool, error) {
	if _, has := smf.skipMap[entry.AbsolutePath]; has {
		return false, nil
	}
	return true, nil
}

// PermissionsFilesFilter filters Entry by ensuring the given file permissions are within the given range.
type PermissionsFilesFilter struct {
	min uint32
	max uint32
}

func NewPermissionsFilesFilter(min, max uint32) PermissionsFilesFilter {
	return PermissionsFilesFilter{min: min, max: max}
}

func (pf PermissionsFilesFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.Mode() < fs.FileMode(pf.min) || entry.FileInfo.Mode() > fs.FileMode(pf.max) {
		return false, nil
	}
	return true, nil
}

// SizeFilesFilter filters Entry by ensuring the given file within the given size range (in bytes).
// Directories are always returned true.
type SizeFilesFilter struct {
	min int64
	max int64
}

func NewSizeFilesFilter(min, max int64) SizeFilesFilter {
	return SizeFilesFilter{min: min, max: max}
}

func (pf SizeFilesFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.IsDir() {
		return true, nil
	} else if entry.FileInfo.Size() < pf.min || entry.FileInfo.Size() > pf.max {
		return false, nil
	}
	return true, nil
}
