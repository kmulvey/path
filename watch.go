package path

import (
	"regexp"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

func WatchDirWithFilter(inputPath string, filter *regexp.Regexp, refreshInterval time.Duration, files chan Entry, shutdown chan struct{}) error {

	defer close(files)
	var uniqFiles = mapset.NewSet[string]()
	var ticker = time.NewTicker(refreshInterval)

	for {
		select {
		case _, open := <-shutdown:
			if !open {
				return nil
			}

		case <-ticker.C:

			var newFiles, err = ListFilesWithFilter(inputPath, filter)
			if err != nil {
				return err
			}

			for _, file := range newFiles {
				if !uniqFiles.Contains(file.AbsolutePath) {
					uniqFiles.Add(file.AbsolutePath)
					files <- file
				}
			}
		}
	}
}
