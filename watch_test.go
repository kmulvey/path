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

	"github.com/stretchr/testify/assert"
)

func TestWatchDirWithFilter(t *testing.T) {
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
