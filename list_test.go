package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/", 3)
	assert.NoError(t, err)
	assert.Equal(t, 9, len(files))

	files, err = List("./notexist/", 3)
	assert.Error(t, err)
	assert.Equal(t, 0, len(files))

	// test filtering out
	files, err = List("./testdata/", 3, NewDirEntitiesFilter())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))

	files, err = List("./testdata/", 3, NewFileEntitiesFilter())
	assert.NoError(t, err)
	assert.Equal(t, 6, len(files))
}
