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

	"github.com/kmulvey/goutils"
)

type File struct {
	DirEntry     fs.DirEntry
	AbsolutePath string
}

// ListFiles is a short cut to list without a regex
func ListFiles(inputPath string) ([]File, error) {
	return ListFilesWithFilter(inputPath, nil)
}

// ListFilesWithFilter un-globs input as well as recursively list all files in the given input
func ListFilesWithFilter(inputPath string, filterRegex *regexp.Regexp) ([]File, error) {
	var allFiles []File

	// expand ~ paths
	if strings.Contains(inputPath, "~") {
		user, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("error getting current user, error: %s", err.Error())
		}
		inputPath = filepath.Join(user.HomeDir, strings.ReplaceAll(inputPath, "~", ""))
	}

	// try un-globing the input
	files, err := filepath.Glob(inputPath)
	if err != nil {
		return nil, fmt.Errorf("could not glob files: %w", err)
	}

	// go through the glob output and expand all dirs
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("could not stat file: %s, err: %w", file, err)
		}

		if stat.IsDir() {
			dirFiles, err := ListDirFiles(file, filterRegex)
			if err != nil {
				return nil, fmt.Errorf("could not list files in dir: %s, err: %w", file, err)
			}
			allFiles = append(allFiles, dirFiles...)
		} else {
			if filterRegex != nil {
				if filterRegex.MatchString(strings.ToLower(stat.Name())) {
					allFiles = append(allFiles, File{AbsolutePath: file, DirEntry: fs.FileInfoToDirEntry(stat)})
				}
			} else {
				allFiles = append(allFiles, File{AbsolutePath: file, DirEntry: fs.FileInfoToDirEntry(stat)})
			}
		}
	}

	return allFiles, nil
}

// ListDirFiles lists every file in a directory (recursive) and
// optionally ignores files given in skipMap
func ListDirFiles(root string, filterRegex *regexp.Regexp) ([]File, error) {
	var allFiles []File
	var files, err = os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("error listing all files in dir: %s, error: %s", root, err.Error())
	}

	for _, file := range files {
		var fullPath = filepath.Join(root, file.Name())

		if file.IsDir() {
			recursiveImages, err := ListDirFiles(fullPath, filterRegex)
			if err != nil {
				return nil, fmt.Errorf("error from recursive call to ListFiles, error: %s", err.Error())
			}
			allFiles = append(allFiles, recursiveImages...)
		} else {
			info, err := file.Info()
			if err != nil {
				return nil, fmt.Errorf("error getting file.Info(), error: %s", err.Error())
			}

			if filterRegex != nil {
				if filterRegex.MatchString(strings.ToLower(file.Name())) {
					allFiles = append(allFiles, File{AbsolutePath: filepath.Join(root, info.Name()), DirEntry: file})
				}
			} else {
				allFiles = append(allFiles, File{AbsolutePath: filepath.Join(root, info.Name()), DirEntry: file})
			}
		}
	}
	return allFiles, nil
}

// DirEntryToString converts a slice of fs.FileInfo to a slice
// of just the files names joined with a given root directory
func DirEntryToString(files []File) []string {
	var fileNames = make([]string, len(files))
	for i, file := range files {
		fileNames[i] = file.AbsolutePath
	}
	return fileNames
}

// FilterFilesByDateRange removes files from the slice if they are not within the given date range.
func FilterFilesByDateRange(files []File, beginTime, endTime time.Time) ([]File, error) {
	for i := len(files) - 1; i >= 0; i-- {
		info, err := files[i].DirEntry.Info()
		if err != nil {
			return nil, err
		}
		if info.ModTime().After(beginTime) && info.ModTime().Before(endTime) {
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
