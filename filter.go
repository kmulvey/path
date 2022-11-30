package path

import (
	"io/fs"
	"regexp"
	"time"

	"github.com/kmulvey/goutils"
)

// FilterEntities removes files from the slice if they are not accepted by the given filter function.
func FilterEntities(files []Entry, filter EntriesFilter) []Entry {
	for i := len(files) - 1; i >= 0; i-- {
		if !filter.filter(files[i]) {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}

// EntitiesFilter interface facilitates filtering entry slices.
type EntriesFilter interface {
	filter(Entry) bool
}

// TrueEntriesFilter always returns true, useful for testing.
type TrueEntriesFilter struct{}

func (nf TrueEntriesFilter) filter(e Entry) bool {
	return true
}

// RegexEntitiesFilter filters fs events by matching file names to a given regex.
type RegexEntitiesFilter struct {
	regex *regexp.Regexp
}

func NewRegexEntitiesFilter(filterRegex *regexp.Regexp) RegexEntitiesFilter {
	return RegexEntitiesFilter{regex: filterRegex}
}

func (rf RegexEntitiesFilter) filter(e Entry) bool {
	return rf.regex.MatchString(e.String())
}

// DateEntitiesFilter filters fs events by matching ensuring ModTime is within the given date range.
type DateEntitiesFilter struct {
	from time.Time
	to   time.Time
}

func NewDateEntitiesFilter(from, to time.Time) DateEntitiesFilter {
	return DateEntitiesFilter{from: from, to: to}
}

func (df DateEntitiesFilter) filter(e Entry) bool {
	if e.FileInfo.ModTime().Before(df.from) || e.FileInfo.ModTime().After(df.to) {
		return true
	}
	return false
}

// SkipMapEntitiesFilter filters fs events by ensuring the given file is NOT within the given map.
type SkipMapEntitiesFilter struct {
	skipMap map[string]struct{}
}

func NewSkipMapEntitiesFilter(skipMap map[string]struct{}) SkipMapEntitiesFilter {
	return SkipMapEntitiesFilter{skipMap: skipMap}
}

func (smf SkipMapEntitiesFilter) filter(e Entry) bool {
	if _, has := smf.skipMap[e.String()]; has {
		return false
	}
	return true
}

// PermissionsEntitiesFilter filters fs events by ensuring the given file permissions are within the given range.
type PermissionsEntitiesFilter struct {
	min uint32
	max uint32
}

func NewPermissionsEntitiesFilter(min, max uint32) PermissionsEntitiesFilter {
	return PermissionsEntitiesFilter{min: min, max: max}
}

func (pf PermissionsEntitiesFilter) filter(e Entry) bool {
	if e.FileInfo.Mode() < fs.FileMode(pf.min) || e.FileInfo.Mode() > fs.FileMode(pf.max) {
		return false
	}
	return true
}

// SizeEntitiesFilter filters fs events by ensuring the given file within the given size range (in bytes).
type SizeEntitiesFilter struct {
	min int64
	max int64
}

func NewSizeEntitiesFilter(min, max int64) SizeEntitiesFilter {
	return SizeEntitiesFilter{min: min, max: max}
}

func (pf SizeEntitiesFilter) filter(e Entry) bool {
	if e.FileInfo.IsDir() {
		return true
	} else if e.FileInfo.Size() < pf.min || e.FileInfo.Size() > pf.max {
		return false
	}
	return true
}
