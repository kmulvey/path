package path

func Contains(input []Entry, needle string) bool {
	for _, entry := range input {
		if entry.AbsolutePath == needle {
			return true
		}
	}
	return false
}

// OnlyNames returns a slice of absolute paths (strings) from a given Entry slice.
func OnlyNames(input []Entry) []string {
	var result = make([]string, len(input))
	for i, entry := range input {
		result[i] = entry.String()
	}
	return result
}
