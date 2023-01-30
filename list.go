package path

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// List recursively lists all files with optional filters. The root directory "inputPath" is excluded from the results.
func List(inputPath string, filters ...ListFilter) ([]Entry, error) {
	var allFiles []Entry

	if strings.TrimSpace(inputPath) == "" {
		return nil, errors.New("inputPath is empty")
	}

	inputPath = filepath.Clean(strings.TrimSpace(inputPath))

	var globFiles, err = unglobInput(inputPath)
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

			abs, err := filepath.Abs(filepath.Clean(path))
			if err != nil {
				return fmt.Errorf("error getting abs path for file: %s, error: %w", path, err)
			}

			var entry = Entry{AbsolutePath: abs, FileInfo: info}

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

// ListFilter interface facilitates filtering of file events.
type ListFilter interface {
	filter(Entry) (bool, error)
}

// RegexListFilter filters Entry by matching file names to a given regex.
type RegexListFilter struct {
	regex *regexp.Regexp
}

func NewRegexListFilter(filterRegex *regexp.Regexp) RegexListFilter {
	return RegexListFilter{regex: filterRegex}
}

func (rf RegexListFilter) filter(entry Entry) (bool, error) {
	return rf.regex.MatchString(entry.String()), nil
}

// DateListFilter filters Entry by matching ensuring ModTime is within the given date range.
type DateListFilter struct {
	from time.Time
	to   time.Time
}

func NewDateListFilter(from, to time.Time) DateListFilter {
	return DateListFilter{from: from, to: to}
}

func (df DateListFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.ModTime().Before(df.from) || entry.FileInfo.ModTime().After(df.to) {
		return false, nil
	}
	return true, nil
}

// SkipMapListFilter filters Entry by ensuring the given file is NOT within the given map.
type SkipMapListFilter struct {
	skipMap map[string]struct{}
}

func NewSkipMapListFilter(skipMap map[string]struct{}) SkipMapListFilter {
	return SkipMapListFilter{skipMap: skipMap}
}

func (smf SkipMapListFilter) filter(entry Entry) (bool, error) {
	if _, has := smf.skipMap[entry.AbsolutePath]; has {
		return false, nil
	}
	return true, nil
}

// PermissionsListFilter filters Entry by ensuring the given file permissions are within the given range.
type PermissionsListFilter struct {
	min uint32
	max uint32
}

func NewPermissionsListFilter(min, max uint32) PermissionsListFilter {
	return PermissionsListFilter{min: min, max: max}
}

func (pf PermissionsListFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.Mode() < fs.FileMode(pf.min) || entry.FileInfo.Mode() > fs.FileMode(pf.max) {
		return false, nil
	}
	return true, nil
}

// SizeListFilter filters Entry by ensuring the given file within the given size range (in bytes).
// Directories are always returned true.
type SizeListFilter struct {
	min int64
	max int64
}

func NewSizeListFilter(min, max int64) SizeListFilter {
	return SizeListFilter{min: min, max: max}
}

func (pf SizeListFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.IsDir() {
		return true, nil
	} else if entry.FileInfo.Size() < pf.min || entry.FileInfo.Size() > pf.max {
		return false, nil
	}
	return true, nil
}

// DirListFilter only returns sub directories of the target.
type DirListFilter struct {
}

func NewDirListFilter() DirListFilter {
	return DirListFilter{}
}

func (df DirListFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.IsDir() {
		return true, nil
	}
	return false, nil
}

// FileListFilter only returns files.
type FileListFilter struct {
}

func NewFileListFilter() FileListFilter {
	return FileListFilter{}
}

func (ff FileListFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.IsDir() {
		return false, nil
	}
	return true, nil
}
