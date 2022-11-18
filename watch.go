package path

import (
	"context"
	"regexp"

	"github.com/fsnotify/fsnotify"
)

func WatchDir(ctx context.Context, inputPath string, filter WatchFilter, files chan Entry) error {

	var errors = make(chan error)
	defer close(files)

	// Create new watcher.
	var watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		defer close(errors)

		for {
			select {
			case <-ctx.Done():
				return

			case event, open := <-watcher.Events:
				if !open {
					return
				}
				var accepted, err = filter.filter(event)
				if err != nil {
					errors <- err
					return
				}
				if accepted {
					if e, err := NewEntry(event.Name); err != nil {
						errors <- err
						return
					} else {
						files <- e
					}
				}

			case err, open := <-watcher.Errors:
				if !open {
					return
				}
				errors <- err
				return
			}
		}
	}()

	// Add a path.
	err = watcher.Add(inputPath)
	if err != nil {
		return err
	}

	return <-errors
}

type WatchFilter interface {
	filter(fsnotify.Event) (bool, error)
}

type NoopFilter struct{}

func (nf NoopFilter) filter(event fsnotify.Event) (bool, error) {
	return true, nil
}

type RegexFilter struct {
	regex *regexp.Regexp
}

func NewRegexFilter(filterRegex *regexp.Regexp) RegexFilter {
	return RegexFilter{regex: filterRegex}
}

func (rf RegexFilter) filter(event fsnotify.Event) (bool, error) {
	return rf.regex.MatchString(event.Name), nil
}
