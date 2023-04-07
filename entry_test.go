package path

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEntry(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 0)
	assert.NoError(t, err)

	TestAbsoultePath(t, entry)
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))

	entry, err = NewEntry("./testdata/*", 1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(entry.Children))
	TestAbsoultePath(t, entry)
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata/*"))

	entry, err = NewEntry("./testdata/", 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(entry.Children))
	TestAbsoultePath(t, entry)
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))

	entry, err = NewEntry("./testdata/", 2, NewDirEntitiesFilter())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entry.Children))
	TestAbsoultePath(t, entry)
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))
}

func TestNewEntryPrivate(t *testing.T) {
	t.Parallel()

	var entry, err = newEntry("./testdata/")
	assert.NoError(t, err)

	TestAbsoultePath(t, entry)
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata"))

	entry, err = newEntry("./testdata/*")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(entry.Children))
	TestAbsoultePath(t, entry)
	assert.True(t, strings.HasSuffix(entry.AbsolutePath, "testdata/*"))
}

func TestPopulateChildren(t *testing.T) {
	t.Parallel()

	var entry = Entry{AbsolutePath: "sdgwgwg/gwegtgh"}
	assert.Error(t, entry.populateChildren(1))

	entry = Entry{AbsolutePath: "./testdata"}
	assert.NoError(t, entry.populateChildren(1))
	assert.Equal(t, 3, len(entry.Children))

	entry = Entry{AbsolutePath: "./testdata"}
	stat, err := os.Lstat("./testdata")
	assert.NoError(t, err)
	entry.FileInfo = stat

	assert.NoError(t, entry.populateChildren(1, NewDirEntitiesFilter()))
	assert.Equal(t, 2, len(entry.Children))
}

func TestCollectChildren(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(entry.Children))

	files, err := entry.Flatten(false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))
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
	assert.Equal(t, 1, len(files))
	assert.True(t, strings.HasPrefix(unglobbedPath, "/"))
	assert.True(t, strings.HasSuffix(unglobbedPath, "testfile"))
}
