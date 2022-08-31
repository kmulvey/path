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

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 8, len(files))

	files, err = ListFiles("./testdata/one/file")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = ListFiles("./doesnotexist")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))

	files, err = ListFiles("./testdata/one/*.mp*")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))

	files, err = ListFiles("./testdata/one/file.mp3")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "./testdata/one/file.mp3", files[0].AbsolutePath)

	files, err = ListFiles("")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))

	// create file in home dir
	user, err := user.Current()
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(user.HomeDir, "pathtestfile"))
	assert.NoError(t, err)

	files, err = ListFiles("~/pathtest*")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = ListFiles("a/b[")
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithFilter(t *testing.T) {
	t.Parallel()

	var suffixRegex = regexp.MustCompile(".*.mp3$")

	files, err := ListFilesWithFilter("./testdata/", suffixRegex)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = ListFilesWithFilter("./testdata/two", suffixRegex)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))

	files, err = ListFilesWithFilter("./testdata/one/file.mp3", suffixRegex)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.True(t, suffixRegex.MatchString(files[0].FileInfo.Name()))

	files, err = ListFilesWithFilter("a/b[", suffixRegex)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithDateFilter(t *testing.T) {
	t.Parallel()

	// set the mod time just in case
	var err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	var fromTime = time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC)
	files, err := ListFilesWithDateFilter("./testdata/", fromTime, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))

	files, err = ListFilesWithDateFilter("a/b[", fromTime, time.Now())
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithMapFilter(t *testing.T) {
	t.Parallel()

	var skipMap = map[string]struct{}{
		"testdata/one/file.mp4": {},
		"testdata/one/file.mp3": {},
	}

	files, err := ListFilesWithMapFilter("./testdata/", skipMap)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(files))

	files, err = ListFilesWithMapFilter("a/b[", skipMap)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithPermissionsFilter(t *testing.T) {
	t.Parallel()

	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	var files, err = ListFilesWithPermissionsFilter("./testdata/", uint32(fs.ModePerm), uint32(fs.ModePerm))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = ListFilesWithPermissionsFilter("a/b[", uint32(fs.ModePerm), uint32(fs.ModePerm))
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestListFilesWithSizeFilter(t *testing.T) {
	t.Parallel()

	var files, err = ListFilesWithSizeFilter("./testdata/", 4100, 6000)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))

	files, err = ListFilesWithSizeFilter("a/b[", 4100, 6000)
	assert.Equal(t, "Error from pre-processing: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
}

func TestOnlyDirs(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)

	files = OnlyDirs(files)
	assert.Equal(t, 3, len(files))
}

func TestOnlyFiles(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)

	files = OnlyFiles(files)
	assert.Equal(t, 5, len(files))
}

func TestOnlyNames(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)

	var strings = OnlyNames(files)
	assert.Equal(t, 8, len(strings))

	for _, str := range strings {
		assert.IsType(t, "", str)
	}
}
