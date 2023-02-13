package path

// List is just a convience function to get a slice of files
func List(inputPath string, levelsDeep int, filters ...EntriesFilter) ([]Entry, error) {

	var entry, err = NewEntry(inputPath, levelsDeep, filters...)
	if err != nil {
		return nil, err
	}

	files, err := entry.Flatten()
	if err != nil {
		return nil, err
	}

	// we filter here again because populateChildren may return dirs that do not match the filter
	for _, file := range files {

		if file.IsDir() {
			for _, fn := range filters {
				var accepted = fn.filter(file)
				if !accepted {
					continue
				}
			}
		}
	}

	return files, nil
}
