package path

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 0)
	assert.NoError(t, err)

	abs, err := filepath.Abs("./testdata/")
	assert.NoError(t, err)

	assert.Equal(t, abs, entry.String())
}

func TestIsDir(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 0)
	assert.NoError(t, err)

	assert.True(t, entry.IsDir())
}

func TestUnglobInput(t *testing.T) {
	t.Parallel()

	var unglobbedPath, files, err = unglobInput("./testdata/*")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))
	assert.Equal(t, "./testdata/*", unglobbedPath)

	unglobbedPath, files, err = unglobInput("a/b[")
	assert.Equal(t, "syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
	assert.Equal(t, "a/b[", unglobbedPath) // no such path exists

	// test a crazy filename
	unglobbedPath, files, err = unglobInput("./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg", unglobbedPath)

}

func TestPopulateChildren(t *testing.T) {
	t.Parallel()

	var entry = Entry{AbsolutePath: "sdgwgwg/gwegtgh"}
	assert.Error(t, entry.populateChildren(1))
}
