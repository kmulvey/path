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

	strings, err := DirEntryToString(files)
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
		"file.mp4": {},
		"file.mp3": {},
	}
	strings, err := FilterFilesBySkipMap(files, skipMap)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(strings))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	for _, str := range strings {
		assert.False(t, suffixRegex.MatchString(str.Name()))
	}
}
