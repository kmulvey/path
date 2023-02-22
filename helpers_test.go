package path

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 3)
	assert.NoError(t, err)

	files, err := entry.Flatten()
	assert.NoError(t, err)
	assert.Equal(t, 9, len(files))

	var mp3Path string

	for _, name := range files {
		if strings.Contains(name.AbsolutePath, ".mp3") {
			mp3Path = name.AbsolutePath
		}
	}

	assert.True(t, Contains(files, mp3Path))

	entry, err = NewEntry("./helpers_test.go", 0)
	assert.NoError(t, err)
	assert.False(t, Contains(files, entry.AbsolutePath))
}

func TestOnlyNames(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 3)
	assert.NoError(t, err)

	files, err := entry.Flatten()
	assert.NoError(t, err)
	assert.Equal(t, 9, len(files))

	for _, name := range OnlyNames(files) {
		assert.True(t, strings.HasPrefix(name, "/"))
	}
}
