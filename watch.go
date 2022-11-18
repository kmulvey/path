package path

import (
	"context"
	"io/fs"
	"regexp"
	"time"

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

//////////////////////////////////////////////////////////////////

// WatchFilter interface facilitates filtering of file events
type WatchFilter interface {
	filter(fsnotify.Event) (bool, error)
}

type NoopFilter struct{}

func (nf NoopFilter) filter(event fsnotify.Event) (bool, error) {
	return true, nil
}

// RegexFilter filters fs events by matching file names to a given regex
type RegexFilter struct {
	regex *regexp.Regexp
}

func NewRegexFilter(filterRegex *regexp.Regexp) RegexFilter {
	return RegexFilter{regex: filterRegex}
}

func (rf RegexFilter) filter(event fsnotify.Event) (bool, error) {
	return rf.regex.MatchString(event.Name), nil
}

// DateFilter filters fs events by matching ensuring ModTime is within the given date range
type DateFilter struct {
	from time.Time
	to   time.Time
}

func NewDateFilter(from, to time.Time) DateFilter {
	return DateFilter{from: from, to: to}
}

func (df DateFilter) filter(event fsnotify.Event) (bool, error) {
	var entry, err = NewEntry(event.Name)
	if err != nil {
		return false, err
	}

	if entry.FileInfo.ModTime().Before(df.from) || entry.FileInfo.ModTime().After(df.to) {
		return true, nil
	}
	return false, nil
}

// SkipMapFilter filters fs events by ensuring the given file is NOT within the given map
type SkipMapFilter struct {
	skipMap map[string]struct{}
}

func NewSkipMapFilter(skipMap map[string]struct{}) SkipMapFilter {
	return SkipMapFilter{skipMap: skipMap}
}

func (smf SkipMapFilter) filter(event fsnotify.Event) (bool, error) {
	var entry, err = NewEntry(event.Name)
	if err != nil {
		return false, err
	}

	if _, has := smf.skipMap[entry.AbsolutePath]; has {
		return false, nil
	}
	return true, nil
}

// PermissionsFilter filters fs events by ensuring the given file permissions are within the given range
type PermissionsFilter struct {
	min uint32
	max uint32
}

func NewPermissionsFilter(min, max uint32) PermissionsFilter {
	return PermissionsFilter{min: min, max: max}
}

func (pf PermissionsFilter) filter(event fsnotify.Event) (bool, error) {
	var entry, err = NewEntry(event.Name)
	if err != nil {
		return false, err
	}

	if entry.FileInfo.Mode() < fs.FileMode(pf.min) || entry.FileInfo.Mode() > fs.FileMode(pf.max) {
		return false, nil
	}
	return true, nil
}

// SizeFilter filters fs events by ensuring the given file within the given size range (in bytes)
type SizeFilter struct {
	min int64
	max int64
}

func NewSizeFilter(min, max int64) SizeFilter {
	return SizeFilter{min: min, max: max}
}

func (pf SizeFilter) filter(event fsnotify.Event) (bool, error) {
	var entry, err = NewEntry(event.Name)
	if err != nil {
		return false, err
	}

	if entry.FileInfo.IsDir() {
		return true, nil
	} else if entry.FileInfo.Size() < pf.min || entry.FileInfo.Size() > pf.max {
		return false, nil
	}
	return true, nil
}
