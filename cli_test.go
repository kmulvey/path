package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCli(t *testing.T) {
	t.Parallel()

	var p = Path{}

	var err = p.Set("./testdata/*")
	assert.NoError(t, err)

	var files = p.Get()
	assert.Equal(t, 5, len(files))
	assert.False(t, Contains(files, "./testdata/"))

	var str = p.String()
	assert.Equal(t, "file file.mp3 file.mp4 file.txt file", str)
}
