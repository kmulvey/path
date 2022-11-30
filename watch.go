package path

import (
	"context"
	"io/fs"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/slices"
)

// WatchEvent is a wrapper for Entry and fsnotify.Op.
type WatchEvent struct {
	Entry
	fsnotify.Op
}

// WatchDir will watch a directory indefinitely for changes and publish them on the given files channel with optional filters.
func WatchDir(ctx context.Context, inputPath string, files chan WatchEvent, filters ...WatchFilter) error {

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
	EventsLoop:
		for {
			select {
			case <-ctx.Done():
				return

			case event, open := <-watcher.Events:
				if !open {
					return
				}
				// try all the filter funcs
				for _, fn := range filters {
					var accepted, err = fn.filter(event)
					if err != nil {
						errors <- err
						return
					}
					if !accepted {
						continue EventsLoop
					}
				}
				if e, err := NewEntry(event.Name); err != nil {
					errors <- err
					return
				} else {
					files <- WatchEvent{Entry: e, Op: event.Op}
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

// WatchFilter interface facilitates filtering of file events.
type WatchFilter interface {
	filter(fsnotify.Event) (bool, error)
}

// TrueWatchFilter always returns true, helpful for tests.
type TrueWatchFilter struct{}

func (nf TrueWatchFilter) filter(event fsnotify.Event) (bool, error) {
	return true, nil
}

// FalseWatchFilter always returns false, helpful for tests.
type FalseWatchFilter struct{}

func (ff FalseWatchFilter) filter(event fsnotify.Event) (bool, error) {
	return false, nil
}

// RegexWatchFilter filters fs events by matching file names to a given regex.
type RegexWatchFilter struct {
	regex *regexp.Regexp
}

func NewRegexWatchFilter(filterRegex *regexp.Regexp) RegexWatchFilter {
	return RegexWatchFilter{regex: filterRegex}
}

func (rf RegexWatchFilter) filter(event fsnotify.Event) (bool, error) {
	return rf.regex.MatchString(event.Name), nil
}

// DateWatchFilter filters fs events by matching ensuring ModTime is within the given date range.
type DateWatchFilter struct {
	from time.Time
	to   time.Time
}

func NewDateWatchFilter(from, to time.Time) DateWatchFilter {
	return DateWatchFilter{from: from, to: to}
}

func (df DateWatchFilter) filter(event fsnotify.Event) (bool, error) {
	var entry, err = NewEntry(event.Name)
	if err != nil {
		return false, err
	}

	if entry.FileInfo.ModTime().Before(df.from) || entry.FileInfo.ModTime().After(df.to) {
		return false, nil
	}
	return true, nil
}

// SkipMapWatchFilter filters fs events by ensuring the given file is NOT within the given map.
type SkipMapWatchFilter struct {
	skipMap map[string]struct{}
}

func NewSkipMapWatchFilter(skipMap map[string]struct{}) SkipMapWatchFilter {
	return SkipMapWatchFilter{skipMap: skipMap}
}

func (smf SkipMapWatchFilter) filter(event fsnotify.Event) (bool, error) {
	var entry, err = NewEntry(event.Name)
	if err != nil {
		return false, err
	}

	if _, has := smf.skipMap[entry.AbsolutePath]; has {
		return false, nil
	}
	return true, nil
}

// PermissionsWatchFilter filters fs events by ensuring the given file permissions are within the given range.
type PermissionsWatchFilter struct {
	min uint32
	max uint32
}

func NewPermissionsWatchFilter(min, max uint32) PermissionsWatchFilter {
	return PermissionsWatchFilter{min: min, max: max}
}

func (pf PermissionsWatchFilter) filter(event fsnotify.Event) (bool, error) {
	var entry, err = NewEntry(event.Name)
	if err != nil {
		return false, err
	}

	if entry.FileInfo.Mode() < fs.FileMode(pf.min) || entry.FileInfo.Mode() > fs.FileMode(pf.max) {
		return false, nil
	}
	return true, nil
}

// SizeWatchFilter filters fs events by ensuring the given file within the given size range (in bytes).
// Directories are always returned true.
type SizeWatchFilter struct {
	min int64
	max int64
}

func NewSizeWatchFilter(min, max int64) SizeWatchFilter {
	return SizeWatchFilter{min: min, max: max}
}

func (pf SizeWatchFilter) filter(event fsnotify.Event) (bool, error) {
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

// OpWatchFilter filters fs events by fsnotify.Op event type.
type OpWatchFilter struct {
	Ops []fsnotify.Op
}

func NewOpWatchFilter(ops ...fsnotify.Op) OpWatchFilter {
	return OpWatchFilter{Ops: ops}
}

func (of OpWatchFilter) filter(event fsnotify.Event) (bool, error) {
	if slices.Contains(of.Ops, event.Op) {
		return true, nil
	}
	return false, nil
}
