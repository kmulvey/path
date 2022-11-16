package path

import (
	"strings"
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
	assert.Equal(t, "testdata/one/file testdata/one/file.mp3 testdata/one/file.mp4 testdata/one/file.txt testdata/two/file", str)

	// this is not the best test but when it runs in ci/cd its hard to predict what the path should look like
	assert.True(t, strings.HasPrefix(p.ComputedPath.AbsolutePath, "/"))

	err = p.Set("~/testdata/*")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(p.ComputedPath.AbsolutePath, "/"))
	assert.Nil(t, p.ComputedPath.FileInfo)

	err = p.Set("./testdata/")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(p.ComputedPath.AbsolutePath, "/"))
	assert.NotNil(t, p.ComputedPath.FileInfo)
}
