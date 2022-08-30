package path

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"testing"

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
	assert.True(t, suffixRegex.MatchString(files[0].DirEntry.Name()))
}

func TestListDirFiles(t *testing.T) {
	t.Parallel()

	files, err := ListDirFiles("", nil)
	assert.Equal(t, "error listing all files in dir: , error: open : no such file or directory", err.Error())
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
