package path

import (
	"fmt"
	"io/fs"
	"regexp"
	"time"

	"golang.org/x/exp/slices"
)

// List recursively lists all files with optional filters. The root directory "inputPath" is excluded from the results.
func List(inputPath string, depth int, filters ...ListFilter) ([]Entry, error) {

	var root, err = NewEntry(inputPath, depth)
	if err != nil {
		return nil, err
	}

	return getChildern(root, depth)
}

func getChildern(entry Entry, depth int, filters ...ListFilter) ([]Entry, error) {

	var entries = entry.Children

	if depth == 1 {

		return entry.Children, nil
	} else {

		for _, child := range entry.Children {

			var newChildren, err = getChildern(child, depth-1)
			if err != nil {
				return nil, err
			}

			// filter out the children we dont need ... sorry kids :(
			for i, childEntry := range newChildren {

				for _, fn := range filters {

					var accepted, err = fn.filter(childEntry)
					if err != nil {
						return nil, fmt.Errorf("error filtering children: %w", err)
					}
					if !accepted {
						newChildren = slices.Delete(newChildren, i, i+1)
					}
				}
			}
			entries = append(entries, newChildren...)
		}
	}

	return entries, nil
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
