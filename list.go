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

// ListAllFiles un-globs input as well as recursivly list all
// files in the given input
func ListAllFiles(inputPath string) ([]os.DirEntry, error) {
	var allFiles []os.DirEntry
	var suffixRegex, err = regexp.Compile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")
	if err != nil {
		return nil, fmt.Errorf("error compiling regex, error: %s", err.Error())
	}

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
			dirFiles, err := ListDirFiles(file)
			if err != nil {
				return nil, fmt.Errorf("could not list files in dir: %s, err: %w", file, err)
			}
			allFiles = append(allFiles, dirFiles...)
		} else {
			if suffixRegex.MatchString(strings.ToLower(file)) {
				allFiles = append(allFiles, fs.FileInfoToDirEntry(stat))
			}
		}
	}

	return allFiles, nil
}

// ListFiles lists every file in a directory (recursive) and
// optionally ignores files given in skipMap
func ListDirFiles(root string) ([]os.DirEntry, error) {
	var allFiles []os.DirEntry
	var files, err = os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("error listing all files in dir: %s, error: %s", root, err.Error())
	}

	suffixRegex, err := regexp.Compile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")
	if err != nil {
		return nil, fmt.Errorf("error compiling regex, error: %s", err.Error())
	}

	for _, file := range files {
		var fullPath = filepath.Join(root, file.Name())
		if file.IsDir() {
			recursiveImages, err := ListDirFiles(fullPath)
			if err != nil {
				return nil, fmt.Errorf("error from recursive call to ListFiles, error: %s", err.Error())
			}
			allFiles = append(allFiles, recursiveImages...)
		} else {
			if suffixRegex.MatchString(strings.ToLower(file.Name())) {
				allFiles = append(allFiles, file)
			}
		}
	}
	return allFiles, nil
}

// FileInfoToString converts a slice of fs.FileInfo to a slice
// of just the files names joined with a given root directory
func FileInfoToString(files []os.DirEntry) ([]string, error) {
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

// FilterFilesByDate removes files from the slice if they were modified
// before the modifiedSince
func FilterFilesByDate(files []os.DirEntry, modifiedSince time.Time) ([]os.DirEntry, error) {
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
func FilterFilesBySkipMap(files []os.DirEntry, skipMap map[string]bool) ([]os.DirEntry, error) {
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

// FilterPNG filters a slice of files to return only pngs
func FilerPNG(files []string) []string {
	var filtered []string
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file), ".png") {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// FilterWEBP filters a slice of files to return only webps
func FilerWEBP(files []string) []string {
	var filtered []string
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file), ".webp") {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// FilterJPG filters a slice of files to return only jpgs
func FilerJPG(files []string) []string {
	var filtered []string
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file), ".jpg") {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// FilterJPEG filters a slice of files to return only jpegs
func FilerJPEG(files []string) []string {
	var filtered []string
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file), ".jpeg") {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// EscapeFilePath escapes spaces in the filepath used for an exec() call
func EscapeFilePath(file string) string {
	var r = strings.NewReplacer(" ", `\ `, "(", `\(`, ")", `\)`, "'", `\'`, "&", `\&`, "@", `\@`)
	return r.Replace(file)
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}
