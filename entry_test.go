package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlyNames(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/")
	assert.NoError(t, err)
	assert.False(t, Contains(files, "./testdata/"))

	var strings = OnlyNames(files)
	assert.Equal(t, 9, len(strings))

	for _, str := range strings {
		assert.IsType(t, "", str)
	}
}

func TestUnglobInput(t *testing.T) {
	t.Parallel()

	var unglobbedPath, files, err = unglobInput("./testdata/*")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))
	assert.Equal(t, "testdata", unglobbedPath)

	unglobbedPath, files, err = unglobInput("a/b[")
	assert.Equal(t, "error unglobbing input, error: syntax error in pattern", err.Error())
	assert.Equal(t, 0, len(files))
	assert.Equal(t, "", unglobbedPath) // no such path exists

}
func TestCrazyFileName(t *testing.T) {
	t.Parallel()

	var unglobbedPath, files, err = unglobInput("./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "testdata", unglobbedPath) // TODO this is not correct
}
