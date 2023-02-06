package path

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlyNames(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/", 2)
	assert.NoError(t, err)
	assert.False(t, Contains(files, "./testdata/"))

	var strings = OnlyNames(files)
	assert.Equal(t, 8, len(strings))

	for _, str := range strings {
		assert.IsType(t, "", str)
	}
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

}
func TestCrazyFileName(t *testing.T) {
	t.Parallel()

	var unglobbedPath, files, err = unglobInput("./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg", unglobbedPath)
}

func TestPopulateChildren(t *testing.T) {
	t.Parallel()

	var entry = Entry{AbsolutePath: "sdgwgwg/gwegtgh"}
	assert.Error(t, entry.populateChildren(1))
}
func TestContains(t *testing.T) {
	t.Parallel()

	var entries, err = List("./testdata/", 2)
	assert.NoError(t, err)

	var mp3Path string

	for _, name := range entries {
		if strings.Contains(name.AbsolutePath, ".mp3") {
			mp3Path = name.AbsolutePath
		}
	}

	assert.True(t, Contains(entries, mp3Path))
}
