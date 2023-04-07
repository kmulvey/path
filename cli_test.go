package path

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fileNotFoundRegex = regexp.MustCompile(`no such file or directory|The system cannot find the path specified`)

func TestCli(t *testing.T) {
	t.Parallel()

	var entry = Entry{}

	var err = entry.Set("./testdata/*")
	assert.NoError(t, err)

	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	var fileMap = map[string]struct{}{
		"two": {},
		"one": {},
		"ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg": {},
	}
	for _, child := range entry.Children {
		delete(fileMap, filepath.Base(child.AbsolutePath))
	}
	assert.Equal(t, 0, len(fileMap))

	var get = entry.Get()
	assert.True(t, prefixRegex.MatchString(get))
	assert.True(t, globSuffixRegex.MatchString(get))

	var str = entry.String()
	assert.Equal(t, get, str)

	err = entry.Set("~/testdata/*")
	assert.True(t, strings.HasPrefix(err.Error(), "error stating file"))
	assert.True(t, fileNotFoundRegex.MatchString(err.Error()))

	assert.Equal(t, "", entry.AbsolutePath)
	assert.Nil(t, entry.FileInfo)

	err = entry.Set("./testdata/")
	assert.NoError(t, err)
	assert.True(t, prefixRegex.MatchString(entry.AbsolutePath))
	assert.NotNil(t, entry.FileInfo)
}
