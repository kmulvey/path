package path

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

func TestWatchDir(t *testing.T) {
	t.Parallel()

	var dir = "./testwatchdir"

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	var files = make(chan Entry)
	var done = make(chan struct{})
	var ctx, cancel = context.WithCancel(context.Background())
	var regexFilter = NewRegexFilter(regexp.MustCompile(".*.txt$"))

	go func() {
		var i int
		for file := range files {
			assert.True(t, strings.HasSuffix(file.AbsolutePath, ".txt"))
			i++
		}
		assert.Equal(t, 2, i)
		close(done)
	}()
	go func() {
		assert.NoError(t, WatchDir(ctx, dir, regexFilter, files))
	}()

	time.Sleep(time.Millisecond * 250) // give time for WatchDir to start up

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte{}, fs.ModePerm))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file1.mp3"), []byte{}, fs.ModePerm))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file2.txt"), []byte{}, fs.ModePerm))

	time.Sleep(time.Millisecond * 250) // give time for WatchDir to process event

	cancel()
	<-done
	assert.NoError(t, os.RemoveAll(dir))
}

func TestNoopFilter(t *testing.T) {
	t.Parallel()

	var noop = NoopFilter{}
	var accpet, err = noop.filter(fsnotify.Event{Name: "/path/to/file"})
	assert.NoError(t, err)
	assert.True(t, accpet)
}

func TestSkipMapFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.txt")
	assert.NoError(t, err)

	var skipMapFilter = NewSkipMapFilter(map[string]struct{}{testFile.AbsolutePath: {}})
	accpet, err := skipMapFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)

	testFileTwo, err := NewEntry("./testdata/one/file.mp3")
	assert.NoError(t, err)

	accpet, err = skipMapFilter.filter(fsnotify.Event{Name: testFileTwo.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)
}

func TestDateFilter(t *testing.T) {
	t.Parallel()

	// set the mod time because in ci/cd the mod time is the time of `git checkout` for the build
	// i.e. "now"
	var err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	testFile, err := NewEntry("./testdata/one/file.mp4")
	assert.NoError(t, err)

	var fromTime = time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC)
	var dateFilter = NewDateFilter(fromTime, time.Now())
	accpet, err := dateFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)

	testFile, err = NewEntry("./testdata/one/file.mp3")
	assert.NoError(t, err)

	accpet, err = dateFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)
}

func TestPermissionsFilter(t *testing.T) {
	t.Parallel()

	// set the perms because the checkout in ci/cd doest match local
	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	var testFile, err = NewEntry("./testdata/one/file.mp3")
	assert.NoError(t, err)

	var permsFilter = NewPermissionsFilter(uint32(fs.ModePerm), uint32(fs.ModePerm))
	accpet, err := permsFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)

	testFile, err = NewEntry("./testdata/one/file.mp4")
	assert.NoError(t, err)

	accpet, err = permsFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)
}

func TestSizeFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.mp4")
	assert.NoError(t, err)

	var sizeFilter = NewSizeFilter(4000, 6000)
	accpet, err := sizeFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)

	testFile, err = NewEntry("./testdata/one/file.mp3")
	assert.NoError(t, err)

	accpet, err = sizeFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)

	testFile, err = NewEntry("./testdata/one")
	assert.NoError(t, err)

	accpet, err = sizeFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)
}
