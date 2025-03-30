package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()

	var files, err = List("./testdata/", 3, false)
	assert.NoError(t, err)
	assert.Len(t, files, 8)

	files, err = List("./testdata/", 3, true)
	assert.NoError(t, err)
	assert.Len(t, files, 9)

	files, err = List("./notexist/", 3, false)
	assert.Error(t, err)
	assert.Empty(t, files)

	// test filtering out
	files, err = List("./testdata/", 3, false, NewDirEntitiesFilter())
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	files, err = List("./testdata/", 3, false, NewFileEntitiesFilter())
	assert.NoError(t, err)
	assert.Len(t, files, 6)
}
