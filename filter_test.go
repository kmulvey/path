package path

import (
	"io/fs"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterFilesSinceDate(t *testing.T) {
	t.Parallel()

	// set the mod time because in ci/cd the mod time is the time of `git checkout` for the build
	// i.e. "now"
	var err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	files, err := ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	var fromTime = time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC)
	files = FilterFilesByDateRange(files, fromTime, time.Now())
	assert.Equal(t, 6, len(files))
}

func TestFilterFilesBySkipMap(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	var skipMap = map[string]struct{}{
		"testdata/one/file.mp4": {},
		"testdata/one/file.mp3": {},
	}
	files = FilterFilesBySkipMap(files, skipMap)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(files))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	for _, str := range files {
		assert.False(t, suffixRegex.MatchString(str.FileInfo.Name()))
	}
}

func TestFilterFilesByRegex(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	files = FilterFilesByRegex(files, suffixRegex)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))
}

func TestFilterFilesByPerms(t *testing.T) {
	t.Parallel()

	// set the perms because the checkout in ci/cd doest match local
	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	files = FilterFilesByPerms(files, uint32(fs.ModePerm), uint32(fs.ModePerm))
	assert.Equal(t, 1, len(files))
}

func TestFilterFilesBySize(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	files = FilterFilesBySize(files, 4000, 6000)
	assert.Equal(t, 1, len(files))
}
