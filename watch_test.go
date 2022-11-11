package path

import (
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
	var shutdown = make(chan struct{})
	var done = make(chan struct{})
	var suffixRegex = regexp.MustCompile(".*.txt$")

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
		assert.NoError(t, WatchDirWithFilter(dir, suffixRegex, time.Millisecond*50, files, shutdown))
	}()

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte{}, fs.ModePerm))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file1.mp3"), []byte{}, fs.ModePerm))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "file2.txt"), []byte{}, fs.ModePerm))

	time.Sleep(time.Millisecond * 250)

	close(shutdown)
	<-done
	assert.NoError(t, os.RemoveAll(dir))
}
