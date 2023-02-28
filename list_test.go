package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/", 3, false)
	assert.NoError(t, err)
	assert.Equal(t, 8, len(files))

	files, err = List("./testdata/", 3, true)
	assert.NoError(t, err)
	assert.Equal(t, 9, len(files))

	files, err = List("./notexist/", 3, false)
	assert.Error(t, err)
	assert.Equal(t, 0, len(files), false)

	// test filtering out
	files, err = List("./testdata/", 3, false, NewDirEntitiesFilter())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(files))

	files, err = List("./testdata/", 3, false, NewFileEntitiesFilter())
	assert.NoError(t, err)
	assert.Equal(t, 6, len(files))
}
