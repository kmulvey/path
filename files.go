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

// preProcessInput expands ~, and un-globs input
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

// ListFiles recursively lists all files
func ListFiles(inputPath string) ([]Entry, error) {
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
			allFiles = append(allFiles, Entry{AbsolutePath: path, FileInfo: info})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}

// ListFilesWithFilter recursively lists all files, filtering the names based on the given regex.
func ListFilesWithFilter(inputPath string, filterRegex *regexp.Regexp) ([]Entry, error) {
	var allFiles []Entry

	var globFiles, err = preProcessInput(inputPath)
	if err != nil {
		return nil, fmt.Errorf("Error from pre-processing: %w", err)
	}

	for _, gf := range globFiles {
		err = filepath.Walk(gf, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Walk error in dir: %q, error: %w", path, err)
			}
			// do not include the root dir
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("Walk error in dir stating file: %s, error: %w", path, err)
			}
			if gf == path && stat.IsDir() {
				return nil
			}
			if filterRegex.MatchString(strings.ToLower(info.Name())) {
				allFiles = append(allFiles, Entry{AbsolutePath: path, FileInfo: info})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}

// ListFilesWithDateFilter recursively lists all files, filtering based on modtime.
func ListFilesWithDateFilter(inputPath string, beginTime, endTime time.Time) ([]Entry, error) {
	var allFiles []Entry

	var globFiles, err = preProcessInput(inputPath)
	if err != nil {
		return nil, fmt.Errorf("Error from pre-processing: %w", err)
	}

	for _, gf := range globFiles {
		err = filepath.Walk(gf, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Walk error in dir: %q, error: %w", path, err)
			}
			// do not include the root dir
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("Walk error in dir stating file: %s, error: %w", path, err)
			}
			if gf == path && stat.IsDir() {
				return nil
			}
			if info.ModTime().After(beginTime) && info.ModTime().Before(endTime) {
				allFiles = append(allFiles, Entry{AbsolutePath: path, FileInfo: info})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}

// ListFilesWithMapFilter recursively lists all files skipping those that exist in the skip map.
func ListFilesWithMapFilter(inputPath string, skipMap map[string]struct{}) ([]Entry, error) {
	var allFiles []Entry

	var globFiles, err = preProcessInput(inputPath)
	if err != nil {
		return nil, fmt.Errorf("Error from pre-processing: %w", err)
	}

	for _, gf := range globFiles {
		err = filepath.Walk(gf, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Walk error in dir: %q, error: %w", path, err)
			}
			// do not include the root dir
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("Walk error in dir stating file: %s, error: %w", path, err)
			}
			if gf == path && stat.IsDir() {
				return nil
			}
			if _, has := skipMap[path]; !has {
				allFiles = append(allFiles, Entry{AbsolutePath: path, FileInfo: info})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}

// ListFilesWithPermissionsFilter recursively lists all files skipping those that are not in the given range, inclusive.
func ListFilesWithPermissionsFilter(inputPath string, min, max uint32) ([]Entry, error) {
	var allFiles []Entry

	var globFiles, err = preProcessInput(inputPath)
	if err != nil {
		return nil, fmt.Errorf("Error from pre-processing: %w", err)
	}

	for _, gf := range globFiles {
		err = filepath.Walk(gf, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Walk error in dir: %q, error: %w", path, err)
			}
			// do not include the root dir
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("Walk error in dir stating file: %s, error: %w", path, err)
			}
			if gf == path && stat.IsDir() {
				return nil
			}
			if info.Mode() >= fs.FileMode(min) && info.Mode() <= fs.FileMode(max) {
				allFiles = append(allFiles, Entry{AbsolutePath: path, FileInfo: info})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}

// ListFilesWithSizeFilter recursively lists all files skipping those that are not in the given range, inclusive.
func ListFilesWithSizeFilter(inputPath string, min, max int64) ([]Entry, error) {
	var allFiles []Entry

	var globFiles, err = preProcessInput(inputPath)
	if err != nil {
		return nil, fmt.Errorf("Error from pre-processing: %w", err)
	}

	for _, gf := range globFiles {
		err = filepath.Walk(gf, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Walk error in dir: %q, error: %w", path, err)
			}
			// do not include the root dir
			stat, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("Walk error in dir stating file: %s, error: %w", path, err)
			}
			if gf == path && stat.IsDir() {
				return nil
			}
			if info.Size() >= min && info.Size() <= max {
				allFiles = append(allFiles, Entry{AbsolutePath: path, FileInfo: info})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allFiles, nil
}
