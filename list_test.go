package path

import (
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreProcessInput(t *testing.T) {
	t.Parallel()

	var files, err = preProcessInput("./testdata/*")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))

	files, err = preProcessInput("a/b[")
	assert.Equal(t, "syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFiles(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))
	assert.False(t, Contains(files, "./testdata/"))
	assert.True(t, files[0].IsDir())

	files, err = List("./testdata/one/file")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = List("./doesnotexist")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))

	files, err = List("./testdata/one/*.mp*")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))

	files, err = List("./testdata/one/file.mp3")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "./testdata/one/file.mp3", files[0].AbsolutePath)

	files, err = List("")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))

	// create file in home dir
	user, err := user.Current()
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(user.HomeDir, "pathtestfile"))
	assert.NoError(t, err)

	files, err = List("~/pathtest*")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = List("a/b[")
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithFilter(t *testing.T) {
	t.Parallel()

	var suffixRegex = regexp.MustCompile(".*.mp3$")
	var suffixRegexFilter = NewRegexListFilter(suffixRegex)

	files, err := List("./testdata/", suffixRegexFilter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = List("./testdata/two", suffixRegexFilter)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))

	files, err = List("./testdata/one/file.mp3", suffixRegexFilter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.True(t, suffixRegex.MatchString(files[0].FileInfo.Name()))

	files, err = List("a/b[", suffixRegexFilter)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithDateFilter(t *testing.T) {
	t.Parallel()

	// set the mod time just in case
	var err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	var fromTimeFilter = NewDateListFilter(time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC), time.Now())
	files, err := List("./testdata/", fromTimeFilter)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	files, err = List("a/b[", fromTimeFilter)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithMapFilter(t *testing.T) {
	t.Parallel()

	var skipMapFilter = NewSkipMapListFilter(map[string]struct{}{
		"testdata/one/file.mp4": {},
		"testdata/one/file.mp3": {},
	})

	files, err := List("./testdata/", skipMapFilter)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	files, err = List("a/b[", skipMapFilter)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithPermissionsFilter(t *testing.T) {
	t.Parallel()

	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	var permsFilter = NewPermissionsListFilter(uint32(fs.ModePerm), uint32(fs.ModePerm))

	var files, err = List("./testdata/", permsFilter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = List("a/b[", permsFilter)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithSizeFilter(t *testing.T) {
	t.Parallel()

	var sizeFilter = NewSizeListFilter(4100, 6000)

	var files, err = List("./testdata/", sizeFilter)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))

	files, err = List("a/b[", sizeFilter)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithDirFilter(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/", NewDirListFilter())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))
}

func TestListFilesWithFileFilter(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/", NewFileListFilter())
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))
}
