package path

import (
	"io/fs"
	"regexp"
	"strings"
	"time"

	"github.com/kmulvey/goutils"
)

// FilterFilesByDateRange removes files from the slice if they are not within the given date range.
func FilterFilesByDateRange(files []Entry, beginTime, endTime time.Time) []Entry {
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].FileInfo.ModTime().Before(beginTime) || files[i].FileInfo.ModTime().After(endTime) {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}

// FilterFilesBySkipMap removes files from the map that are also in the skipMap.
func FilterFilesBySkipMap(files []Entry, skipMap map[string]struct{}) []Entry {
	for i := len(files) - 1; i >= 0; i-- {
		if _, has := skipMap[files[i].AbsolutePath]; has {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}

// FilterFilesByRegex removes files from the slice if they do not match the regex.
func FilterFilesByRegex(files []Entry, filterRegex *regexp.Regexp) []Entry {
	for i := len(files) - 1; i >= 0; i-- {
		if !filterRegex.MatchString(strings.ToLower(files[i].AbsolutePath)) {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}

// FilterFilesByPerms removes files from the slice if they are not in the given range.
func FilterFilesByPerms(files []Entry, min, max uint32) []Entry {
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].FileInfo.Mode() < fs.FileMode(min) || files[i].FileInfo.Mode() > fs.FileMode(max) {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}

// FilterFilesBySize removes files from the slice if they are not in the given range.
// Ignores dirs.
func FilterFilesBySize(files []Entry, min, max int64) []Entry {
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].FileInfo.IsDir() {
			files = goutils.RemoveElementFromArray(files, i)
		} else if files[i].FileInfo.Size() < min || files[i].FileInfo.Size() > max {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}
