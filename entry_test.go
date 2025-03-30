package path

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var prefixRegex = regexp.MustCompile(`^\/|^[A-Z]:\\`)
var globSuffixRegex = regexp.MustCompile(`\/\*$|\\\*$`)

func TestNewEntry(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 0)
	assert.NoError(t, err)

	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))

	entry, err = NewEntry("./testdata/*", 1)
	assert.NoError(t, err)
	assert.Len(t, entry.Children, 3)
	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	assert.True(t, globSuffixRegex.MatchString(entry.AbsolutePath))

	entry, err = NewEntry("./testdata/", 2)
	assert.NoError(t, err)
	assert.Len(t, entry.Children, 3)
	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))

	entry, err = NewEntry("./testdata/", 2, NewDirEntitiesFilter())
	assert.NoError(t, err)
	assert.Len(t, entry.Children, 2)
	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))
}

func TestNewEntryPrivate(t *testing.T) {
	t.Parallel()

	var entry, err = newEntry("./testdata/")
	assert.NoError(t, err)

	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))

	entry, err = newEntry("./testdata/*")
	assert.NoError(t, err)
	assert.Len(t, entry.Children, 3)
	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	assert.True(t, globSuffixRegex.MatchString(entry.AbsolutePath))
}

func TestPopulateChildren(t *testing.T) {
	t.Parallel()

	var entry = Entry{AbsolutePath: "sdgwgwg/gwegtgh"}
	assert.Error(t, entry.populateChildren(1))

	entry = Entry{AbsolutePath: "./testdata"}
	assert.NoError(t, entry.populateChildren(1))
	assert.Len(t, entry.Children, 3)

	entry = Entry{AbsolutePath: "./testdata"}
	stat, err := os.Lstat("./testdata")
	assert.NoError(t, err)
	entry.FileInfo = stat

	assert.NoError(t, entry.populateChildren(1, NewDirEntitiesFilter()))
	assert.Len(t, entry.Children, 2)
}

func TestCollectChildren(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 1)
	assert.NoError(t, err)
	assert.Len(t, entry.Children, 3)

	files, err := entry.Flatten(false)
	assert.NoError(t, err)
	assert.Len(t, files, 3)
}

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
	assert.Len(t, files, 3)
	assert.Equal(t, "./testdata/*", unglobbedPath)

	unglobbedPath, files, err = unglobInput("a/b[")
	assert.Equal(t, "failed to glob input path a/b[: syntax error in pattern", err.Error())
	assert.Empty(t, files)
	assert.Empty(t, unglobbedPath) // no such path exists

	// test a crazy filename
	unglobbedPath, files, err = unglobInput("./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg")
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg", unglobbedPath)

	// test ~
	user, err := user.Current()
	assert.NoError(t, err)
	var testFile = filepath.Join(user.HomeDir, "testfile")
	defer assert.NoError(t, os.RemoveAll(testFile))

	if _, err := os.Lstat(testFile); errors.Is(err, os.ErrNotExist) {
		f, createErr := os.Create(testFile)
		assert.NoError(t, createErr)
		assert.NoError(t, f.Close())
	}

	unglobbedPath, files, err = unglobInput("~/testfile")
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.True(t, prefixRegex.MatchString(unglobbedPath))
	assert.True(t, strings.HasSuffix(unglobbedPath, "testfile"))
}
