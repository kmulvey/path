package path

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

func TestWatchDir(t *testing.T) {
	t.Parallel()

	var dir = "./testwatchdir"

	if _, err := os.Lstat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	var files = make(chan WatchEvent)
	var done = make(chan struct{})
	var ctx, cancel = context.WithCancel(context.Background())
	var regexFilter = NewRegexWatchFilter(regexp.MustCompile(".*.txt$"))

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
		var errors = make(chan error)
		go func() {
			for err := range errors {
				assert.NoError(t, err)
			}
		}()

		WatchDir(ctx, dir, 0, false, files, errors, regexFilter)
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

func TestWatchDirRecursive(t *testing.T) {
	t.Parallel()

	var dir = "./testwatchdirrecursive"

	if _, err := os.Lstat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(filepath.Join(dir, "one"), os.ModePerm)
		assert.NoError(t, err)

		err = os.MkdirAll(filepath.Join(dir, "two"), os.ModePerm)
		assert.NoError(t, err)

	}

	var files = make(chan WatchEvent)
	var done = make(chan struct{})
	var ctx, cancel = context.WithCancel(context.Background())
	var regexFilter = NewRegexWatchFilter(regexp.MustCompile(".*.txt$"))

	go func() {
		var dirOne, dirTwo int
		for file := range files {
			assert.True(t, strings.HasSuffix(file.AbsolutePath, ".txt"))
			if strings.Contains(file.AbsolutePath, "testwatchdirrecursive/one") || strings.Contains(file.AbsolutePath, `testwatchdirrecursive\one`) {
				dirOne++
			} else if strings.Contains(file.AbsolutePath, "testwatchdirrecursive/two") || strings.Contains(file.AbsolutePath, `testwatchdirrecursive\two`) {
				dirTwo++
			}
		}
		assert.Equal(t, 1, dirOne)
		assert.Equal(t, 1, dirTwo)
		close(done)
	}()
	go func() {
		var errors = make(chan error)
		go func() {
			for err := range errors {
				assert.NoError(t, err)
			}
		}()

		WatchDir(ctx, dir, 2, false, files, errors, regexFilter)
	}()

	time.Sleep(time.Millisecond * 250) // give time for WatchDir to start up

	assert.NoError(t, os.WriteFile(filepath.Join(filepath.Join(dir, "two"), "file1.txt"), []byte{}, fs.ModePerm))
	assert.NoError(t, os.WriteFile(filepath.Join(filepath.Join(dir, "two"), "file1.mp3"), []byte{}, fs.ModePerm))
	assert.NoError(t, os.WriteFile(filepath.Join(filepath.Join(dir, "one"), "file2.txt"), []byte{}, fs.ModePerm))

	time.Sleep(time.Millisecond * 250) // give time for WatchDir to process event

	cancel()
	<-done
	assert.NoError(t, os.RemoveAll(dir))
}

func TestSkipMapWatchFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.txt", 0)
	assert.NoError(t, err)

	var skipMapFilter = NewSkipMapWatchFilter(map[string]struct{}{testFile.AbsolutePath: {}})
	accpet, err := skipMapFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)

	testFileTwo, err := NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)

	accpet, err = skipMapFilter.filter(fsnotify.Event{Name: testFileTwo.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)

	accpet, err = skipMapFilter.filter(fsnotify.Event{Name: "filenotexists"})
	assert.Error(t, err)
	assert.False(t, accpet)
}

func TestDateWatchFilter(t *testing.T) {
	t.Parallel()

	// set the mod time because in ci/cd the mod time is the time of `git checkout` for the build
	// i.e. "now"
	var err = os.Chtimes("./testdata/one/file.mp4", time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC), time.Date(2022, 06, 01, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)

	testFile, err := NewEntry("./testdata/one/file.mp4", 0)
	assert.NoError(t, err)

	var fromTime = time.Date(2022, 07, 01, 0, 0, 0, 0, time.UTC)
	var dateFilter = NewDateWatchFilter(fromTime, time.Now())
	accpet, err := dateFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)

	testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)

	accpet, err = dateFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)

	accpet, err = dateFilter.filter(fsnotify.Event{Name: "filenotexists"})
	assert.Error(t, err)
	assert.False(t, accpet)
}

func TestPermissionsWatchFilter(t *testing.T) {
	t.Parallel()

	// set the perms because the checkout in ci/cd doest match local
	assert.NoError(t, os.Chmod("./testdata/one/file.mp3", fs.ModePerm))

	var testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)

	var permsFilter = NewPermissionsWatchFilter(uint32(fs.ModePerm), uint32(fs.ModePerm))
	accpet, err := permsFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	if runtime.GOOS != "windows" { // i give up trying to figure out how windows does perms
		assert.True(t, accpet)
	}

	testFile, err = NewEntry("./testdata/one/file.mp4", 0)
	assert.NoError(t, err)

	accpet, err = permsFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)
}

func TestSizeWatchFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.mp4", 0)
	assert.NoError(t, err)

	var sizeFilter = NewSizeWatchFilter(4000, 6000)
	accpet, err := sizeFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)

	testFile, err = NewEntry("./testdata/one/file.mp3", 0)
	assert.NoError(t, err)

	accpet, err = sizeFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.False(t, accpet)

	testFile, err = NewEntry("./testdata/one", 0)
	assert.NoError(t, err)

	accpet, err = sizeFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath})
	assert.NoError(t, err)
	assert.True(t, accpet)

	accpet, err = sizeFilter.filter(fsnotify.Event{Name: "filenotexists"})
	assert.Error(t, err)
	assert.False(t, accpet)
}

func TestOpWatchFilter(t *testing.T) {
	t.Parallel()

	var testFile, err = NewEntry("./testdata/one/file.mp4", 0)
	assert.NoError(t, err)

	var opFilter = NewOpWatchFilter(fsnotify.Create)
	accpet, err := opFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath, Op: fsnotify.Create})
	assert.NoError(t, err)
	assert.True(t, accpet)

	accpet, err = opFilter.filter(fsnotify.Event{Name: testFile.AbsolutePath, Op: fsnotify.Remove})
	assert.NoError(t, err)
	assert.False(t, accpet)
}

func TestDirWatchFilter(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/one", 1)
	assert.NoError(t, err)

	var dirFilter = NewDirWatchFilter()
	accpet, err := dirFilter.filter(entry)
	assert.NoError(t, err)
	assert.True(t, accpet)

	entry, err = NewEntry("./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg", 1)
	assert.NoError(t, err)

	accpet, err = dirFilter.filter(entry)
	assert.NoError(t, err)
	assert.False(t, accpet)
}

func TestFileWatchFilter(t *testing.T) {
	t.Parallel()

	var entry, err = NewEntry("./testdata/one", 1)
	assert.NoError(t, err)

	var dirFilter = NewFileWatchFilter()
	accpet, err := dirFilter.filter(entry)
	assert.NoError(t, err)
	assert.False(t, accpet)

	entry, err = NewEntry("./testdata/ogCGs91VSA5FBjJdgE8eeLSngbebPXyDCICZ7I~tplv-f5insbecw7-1 720 720.jpg", 1)
	assert.NoError(t, err)

	accpet, err = dirFilter.filter(entry)
	assert.NoError(t, err)
	assert.True(t, accpet)
}
