package path

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListFiles(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))

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

	// create file in home dir
	user, err := user.Current()
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(user.HomeDir, "pathtestfile"))
	assert.NoError(t, err)

	files, err = ListFiles("~/pathtest*")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
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
}

func TestDirEntryToString(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))

	var strings = DirEntryToString(files)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(strings))
	for _, str := range strings {
		assert.IsType(t, "", str)
	}
}

func TestFilterFilesSinceDate(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))

	// set the mod time just in case
	err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	var fromTime = time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC)
	strings, err := FilterFilesSinceDate(files, fromTime)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(strings))
}

func TestFilterFilesBySkipMap(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))

	var skipMap = map[string]struct{}{
		"testdata/one/file.mp4": {},
		"testdata/one/file.mp3": {},
	}
	files = FilterFilesBySkipMap(files, skipMap)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	for _, str := range files {
		assert.False(t, suffixRegex.MatchString(str.DirEntry.Name()))
	}
}

func TestFilterFilesByRegex(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	files = FilterFilesByRegex(files, suffixRegex)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))

	for _, str := range files {
		assert.True(t, suffixRegex.MatchString(str.DirEntry.Name()))
	}
}
