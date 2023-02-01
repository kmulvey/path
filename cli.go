package path

// Get fulfils the flag.Getter interface https://pkg.go.dev/flag#Getter.
func (e *Entry) Get() string {
	return e.AbsolutePath
}

// Set fulfils the flag.Value interface https://pkg.go.dev/flag#Value.
func (e *Entry) Set(s string) error {

	var entry, err = NewEntry(s, 1)
	e.AbsolutePath = entry.AbsolutePath
	e.Children = entry.Children
	e.FileInfo = entry.FileInfo

	return err
}
