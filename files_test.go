package path

import (
	"fmt"
	"io/fs"
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
}

func TestListFilesWithPermissionsFilter(t *testing.T) {
	t.Parallel()

	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	var files, err = ListFilesWithPermissionsFilter("./testdata/", uint32(fs.ModePerm), uint32(fs.ModePerm))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
}

func TestListFilesWithSizeFilter(t *testing.T) {
	t.Parallel()

	var files, err = ListFilesWithSizeFilter("./testdata/", 4000, 6000)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	for _, f := range files {
		fmt.Println(f.AbsolutePath, f.FileInfo.Size())
	}
}
