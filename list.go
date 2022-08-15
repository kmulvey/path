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

func ListFiles(inputPath string) ([]os.DirEntry, error) {
	return ListFilesWithFilter(inputPath, nil)
}

// ListFilesWithFilter un-globs input as well as recursively list all
// files in the given input
func ListFilesWithFilter(inputPath string, filterRegex *regexp.Regexp) ([]os.DirEntry, error) {
	var allFiles []os.DirEntry

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
					allFiles = append(allFiles, fs.FileInfoToDirEntry(stat))
				}
			} else {
				allFiles = append(allFiles, fs.FileInfoToDirEntry(stat))
			}
		}
	}

	return allFiles, nil
}

// ListFiles lists every file in a directory (recursive) and
// optionally ignores files given in skipMap
func ListDirFiles(root string, filterRegex *regexp.Regexp) ([]os.DirEntry, error) {
	var allFiles []os.DirEntry
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
			if filterRegex != nil {
				if filterRegex.MatchString(strings.ToLower(file.Name())) {
					allFiles = append(allFiles, file)
				}
			} else {
				allFiles = append(allFiles, file)
			}
		}
	}
	return allFiles, nil
}

// DirEntryToString converts a slice of fs.FileInfo to a slice
// of just the files names joined with a given root directory
func DirEntryToString(files []os.DirEntry) ([]string, error) {
	var fileNames = make([]string, len(files))
	for i, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}
		fileNames[i] = info.Name()
	}
	return fileNames, nil
}

// FilterFilesSinceDate removes files from the slice if they were modified
// before the modifiedSince
func FilterFilesSinceDate(files []os.DirEntry, modifiedSince time.Time) ([]os.DirEntry, error) {
	for i := len(files) - 1; i >= 0; i-- {
		info, err := files[i].Info()
		if err != nil {
			return nil, err
		}
		if info.ModTime().Before(modifiedSince) {
			files = remove(files, i)
		}
	}
	return files, nil
}

// FilterFilesBySkipMap removes files from the map that are also in the skipMap
func FilterFilesBySkipMap(files []os.DirEntry, skipMap map[string]struct{}) ([]os.DirEntry, error) {
	for i := len(files) - 1; i >= 0; i-- {
		info, err := files[i].Info()
		if err != nil {
			return nil, err
		}
		if _, has := skipMap[info.Name()]; has {
			files = remove(files, i)
		}
	}
	return files, nil
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}
