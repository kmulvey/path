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

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 8, len(files))

	// set the mod time just in case
	err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	var fromTime = time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC)
	files, err = FilterFilesByDateRange(files, fromTime, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, 7, len(files))
}

func TestFilterFilesBySkipMap(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 8, len(files))

	var skipMap = map[string]struct{}{
		"testdata/one/file.mp4": {},
		"testdata/one/file.mp3": {},
	}
	files = FilterFilesBySkipMap(files, skipMap)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(files))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	for _, str := range files {
		assert.False(t, suffixRegex.MatchString(str.FileInfo.Name()))
	}
}

func TestFilterFilesByRegex(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 8, len(files))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	files = FilterFilesByRegex(files, suffixRegex)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))

	for _, str := range files {
		assert.True(t, suffixRegex.MatchString(str.FileInfo.Name()))
	}
}

func TestFilterFilesByPerms(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 8, len(files))

	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	files, err = FilterFilesByPerms(files, uint32(fs.ModePerm), uint32(fs.ModePerm))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
}

func TestFilterFilesBySize(t *testing.T) {
	t.Parallel()

	var files, err = ListFiles("./testdata/")
	assert.NoError(t, err)
	assert.Equal(t, 8, len(files))

	files, err = FilterFilesBySize(files, 4000, 6000)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
}