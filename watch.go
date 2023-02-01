package path

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
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
func WatchDir(ctx context.Context, inputPath string, recursiveDepth int, files chan WatchEvent, errors chan error, filters ...WatchFilter) {

	inputPath = filepath.Clean(strings.TrimSpace(inputPath))

	var inputEntry, err = NewEntry(inputPath, recursiveDepth)
	if err != nil {
		errors <- fmt.Errorf("error with inputPath: %w", err)
		return
	}

	defer close(files)
	defer close(errors)

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		errors <- fmt.Errorf("error creating NewWatcher: %w", err)
		return
	}
	defer watcher.Close()

	// Start listening for events.
	var wait = make(chan struct{})
	go func() {
		defer close(wait)

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
					}
					if !accepted {
						continue EventsLoop
					}
				}
				if e, err := NewEntry(event.Name, recursiveDepth); err != nil {
					errors <- err
				} else {
					files <- WatchEvent{Entry: e, Op: event.Op}
				}

			case err, open := <-watcher.Errors:
				if !open {
					return
				}
				errors <- err
			}
		}
	}()

	var entries []Entry
	if recursiveDepth > 0 {
		entries, err = List(inputPath, NewDirListFilter())
		if err != nil {
			errors <- fmt.Errorf("error adding path to watcher: %w", err)
			return
		}
		entries = append(entries, inputEntry)
	} else {
		entries = []Entry{inputEntry}
	}

	// Add paths.
	for _, dir := range entries {
		err = watcher.Add(dir.AbsolutePath)
		if err != nil {
			errors <- fmt.Errorf("error adding path to watcher: %w", err)
			return
		}
	}

	<-wait
}

//////////////////////////////////////////////////////////////////

// WatchFilter interface facilitates filtering of file events.
type WatchFilter interface {
	filter(fsnotify.Event) (bool, error)
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
	var entry, err = NewEntry(event.Name, 0)
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
	var entry, err = NewEntry(event.Name, 0)
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
	var entry, err = NewEntry(event.Name, 0)
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
	var entry, err = NewEntry(event.Name, 0)
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

// DirWatchFilter only returns sub directories of the target.
type DirWatchFilter struct{}

func NewDirWatchFilter() DirWatchFilter {
	return DirWatchFilter{}
}

func (df DirWatchFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.IsDir() {
		return true, nil
	}
	return false, nil
}

// FileWatchFilter only returns files.
type FileWatchFilter struct{}

func NewFileWatchFilter() FileWatchFilter {
	return FileWatchFilter{}
}

func (ff FileWatchFilter) filter(entry Entry) (bool, error) {
	if entry.FileInfo.IsDir() {
		return false, nil
	}
	return true, nil
}
