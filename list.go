package path

import (
	"os"
)

// List is just a convience function to get a slice of files
func List(inputPath string, levelsDeep uint8, includeRoot bool, filters ...EntriesFilter) ([]Entry, error) {

	var entry, err = NewEntry(inputPath, levelsDeep, filters...)
	if err != nil {
		return nil, err
	}

	files, err := entry.Flatten(includeRoot)
	if err != nil {
		return nil, err
	}

	var filteredFiles []Entry

	// we filter here again because populateChildren may return dirs that do not match the filter
FileLoop:
	for _, file := range files {

		if file.IsDir() || file.FileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			for _, fn := range filters {
				var accepted = fn.filter(file)
				if !accepted {
					continue FileLoop
				}
			}
		}
		filteredFiles = append(filteredFiles, file)
	}

	return filteredFiles, nil
}
