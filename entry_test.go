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
	assert.Equal(t, 7, len(strings))

	for _, str := range strings {
		assert.IsType(t, "", str)
	}
}
