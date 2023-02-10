package path

import (
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterEntities(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/", 3)
	assert.NoError(t, err)

	files, err := entry.Flatten()
	assert.Equal(t, 8, len(files))

	var skipMap = make(map[string]struct{})

	for _, file := range files {
		if strings.Contains(file.AbsolutePath, ".mp3") || strings.Contains(file.AbsolutePath, ".mp4") {
			skipMap[file.AbsolutePath] = struct{}{}
		}
	}

	var skipMapFilter = NewSkipMapEntitiesFilter(skipMap)

	files = FilterEntities(files, skipMapFilter)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(files))

	var suffixRegex = regexp.MustCompile(".*.mp3$|.*.mp4$")
	for _, str := range files {
		assert.False(t, suffixRegex.MatchString(str.FileInfo.Name()))
	}
}
func TestRegexEntitiesFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.txt", 0)
	assert.NoError(t, err)

	var regexFilter = NewRegexEntitiesFilter(regexp.MustCompile(".*.txt$"))
	assert.True(t, regexFilter.filter(testFile))

	testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)
	assert.False(t, regexFilter.filter(testFile))
}

func TestDateEntitiesFilter(t *testing.T) {
	t.Parallel()

	// set the mod time because in ci/cd the mod time is the time of `git checkout` for the build
	// i.e. "now"
	var err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	testFile, err := NewEntry("./testdata/one/file.mp4", 0)
	assert.NoError(t, err)

	var fromTime = time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC)
	var dateFilter = NewDateEntitiesFilter(fromTime, time.Now())
	assert.False(t, dateFilter.filter(testFile))

	testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)

	assert.True(t, dateFilter.filter(testFile))
}

func TestSkipMapEntitiesFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.txt", 0)
	assert.NoError(t, err)

	var skipMapFilter = NewSkipMapEntitiesFilter(map[string]struct{}{testFile.AbsolutePath: {}})
	assert.False(t, skipMapFilter.filter(testFile))

	testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)

	assert.True(t, skipMapFilter.filter(testFile))
}

func TestPermissionsEntitiesFilter(t *testing.T) {
	t.Parallel()

	// set the perms because the checkout in ci/cd doest match local
	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	var testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)

	var permsFilter = NewPermissionsEntitiesFilter(uint32(fs.ModePerm), uint32(fs.ModePerm))
	assert.True(t, permsFilter.filter(testFile))

	testFile, err = NewEntry("./testdata/one/file.mp4", 0)
	assert.NoError(t, err)

	assert.False(t, permsFilter.filter(testFile))
}

func TestSizeEntitiesFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.mp4", 0)
	assert.NoError(t, err)

	var sizeFilter = NewSizeEntitiesFilter(4000, 6000)
	assert.True(t, sizeFilter.filter(testFile))

	testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)
	assert.False(t, sizeFilter.filter(testFile))

	testFile, err = NewEntry("./testdata/one", 0)
	assert.NoError(t, err)
	assert.True(t, sizeFilter.filter(testFile))
}

func TestDirEntitiesFilter(t *testing.T) {
	t.Parallel()

	var filter = NewDirEntitiesFilter()
	var testFile, err = NewEntry("./testdata/one", 0)
	assert.NoError(t, err)
	assert.True(t, filter.filter(testFile))

	testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)
	assert.False(t, filter.filter(testFile))
}

func TestFileEntitiesFilter(t *testing.T) {
	t.Parallel()

	var filter = NewFileEntitiesFilter()
	var testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)
	assert.True(t, filter.filter(testFile))

	testFile, err = NewEntry("./testdata/one/", 0)
	assert.NoError(t, err)
	assert.False(t, filter.filter(testFile))
}
