package path

import (
	"io/fs"
	"regexp"
	"strings"
	"time"

	"github.com/kmulvey/goutils"
)

// FilterFilesByDateRange removes files from the slice if they are not within the given date range.
func FilterFilesByDateRange(files []File, beginTime, endTime time.Time) ([]File, error) {
	for i := len(files) - 1; i >= 0; i-- {
		info, err := files[i].DirEntry.Info()
		if err != nil {
			return nil, err
		}
		if info.ModTime().Before(beginTime) || info.ModTime().After(endTime) {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files, nil
}

// FilterFilesBySkipMap removes files from the map that are also in the skipMap.
func FilterFilesBySkipMap(files []File, skipMap map[string]struct{}) []File {
	for i := len(files) - 1; i >= 0; i-- {
		if _, has := skipMap[files[i].AbsolutePath]; has {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}

// FilterFilesByRegex removes files from the slice if they do not match the regex.
func FilterFilesByRegex(files []File, filterRegex *regexp.Regexp) []File {
	for i := len(files) - 1; i >= 0; i-- {
		if !filterRegex.MatchString(strings.ToLower(files[i].AbsolutePath)) {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files
}

// FilterFilesByPerms removes files from the slice if they are not in the given range.
func FilterFilesByPerms(files []File, min, max uint32) ([]File, error) {
	for i := len(files) - 1; i >= 0; i-- {
		info, err := files[i].DirEntry.Info()
		if err != nil {
			return nil, err
		}
		if info.Mode() < fs.FileMode(min) || info.Mode() > fs.FileMode(max) {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files, nil
}

// FilterFilesBySize removes files from the slice if they are not in the given range.
func FilterFilesBySize(files []File, min, max int64) ([]File, error) {
	for i := len(files) - 1; i >= 0; i-- {
		info, err := files[i].DirEntry.Info()
		if err != nil {
			return nil, err
		}
		if info.Size() < min || info.Size() > max {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files, nil
}

/*
// OnlyDirs removes files from the slice if they do not match the regex.
func OnlyDirs(files []File) ([]File, error) {
	for i := len(files) - 1; i >= 0; i-- {
		info, err := files[i].DirEntry.Info()
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = goutils.RemoveElementFromArray(files, i)
		}
	}
	return files, nil
}
*/
