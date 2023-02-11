package path

// List is just a convience function to get a slice of files
func List(inputPath string, levelsDeep int, filters ...EntriesFilter) ([]Entry, error) {

	var entry, err = NewEntry(inputPath, levelsDeep, filters...)
	if err != nil {
		return nil, err
	}

	return entry.Flatten()
}
