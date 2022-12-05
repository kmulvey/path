package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlyDirs(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/")
	assert.NoError(t, err)
	assert.False(t, Contains(files, "./testdata/"))

	var dirs = OnlyDirs(files)
	assert.Equal(t, 2, len(dirs))
	assert.True(t, Contains(dirs, "testdata/one"))
	assert.True(t, Contains(dirs, "testdata/two"))
	assert.False(t, Contains(dirs, "testdata/one/file.mp3"))
}

func TestOnlyFiles(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/")
	assert.NoError(t, err)
	assert.False(t, Contains(files, "./testdata/"))

	var dirs = OnlyFiles(files)
	assert.Equal(t, 5, len(dirs))
	assert.False(t, Contains(dirs, "testdata/one"))
	assert.False(t, Contains(dirs, "testdata/two"))
	assert.True(t, Contains(dirs, "testdata/one/file.mp3"))
}

func TestOnlyNames(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/")
	assert.NoError(t, err)
	assert.False(t, Contains(files, "./testdata/"))

	var strings = OnlyNames(files)
	assert.Equal(t, 7, len(strings))

	for _, str := range strings {
		assert.IsType(t, "", str)
	}
}
