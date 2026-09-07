package daemon

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const defaultDebounceDuration = 2500 * time.Millisecond

// FolderWatcher monitors a folder recursively for user-initiated file modifications,
// filtering out transient editor/lock files and debouncing bursts of events.
type FolderWatcher struct {
	rootPath string
	watcher  *fsnotify.Watcher
	debounce time.Duration
	changeCh chan struct{}

	mu         sync.Mutex
	suppressed bool
	timer      *time.Timer
}

// NewFolderWatcher creates and initializes a recursive file watcher on rootPath.
func NewFolderWatcher(rootPath string, debounce time.Duration) (*FolderWatcher, error) {
	if debounce <= 0 {
		debounce = defaultDebounceDuration
	}
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FolderWatcher{
		rootPath: rootPath,
		watcher:  fsw,
		debounce: debounce,
		changeCh: make(chan struct{}, 1),
	}

	if err := fw.addRecursive(rootPath); err != nil {
		_ = fsw.Close()
		return nil, err
	}

	return fw, nil
}

// Events returns the channel that receives a signal when a debounced change has stabilized.
func (fw *FolderWatcher) Events() <-chan struct{} {
	return fw.changeCh
}

// SetSuppressed controls whether incoming filesystem events are ignored (e.g. during sync writes).
func (fw *FolderWatcher) SetSuppressed(suppressed bool) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.suppressed = suppressed
	if suppressed && fw.timer != nil {
		fw.timer.Stop()
		fw.timer = nil
	}
}

// Drain clears any pending change notifications from the channel.
func (fw *FolderWatcher) Drain() {
	for {
		select {
		case <-fw.changeCh:
		default:
			return
		}
	}
}

// Close closes the underlying filesystem watcher and stops timers.
func (fw *FolderWatcher) Close() error {
	fw.mu.Lock()
	if fw.timer != nil {
		fw.timer.Stop()
		fw.timer = nil
	}
	fw.mu.Unlock()
	return fw.watcher.Close()
}

// Start begins listening to filesystem events in the background until ctx is canceled.
func (fw *FolderWatcher) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			_ = err
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)
		}
	}
}

func (fw *FolderWatcher) handleEvent(event fsnotify.Event) {
	name := event.Name
	base := filepath.Base(name)

	// Ignore hidden files and directories
	if isIgnoredFile(base) {
		return
	}

	// Dynamically watch newly created subdirectories
	if event.Op&fsnotify.Create != 0 {
		if fi, err := os.Stat(name); err == nil && fi.IsDir() {
			_ = fw.addRecursive(name)
		}
	}

	// Only trigger on write/create/remove/rename operations
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return
	}

	fw.mu.Lock()
	defer fw.mu.Unlock()

	// If currently suppressed (sync writing local files), drop the event to avoid echo loops
	if fw.suppressed {
		return
	}

	// Reset or create debounce timer
	if fw.timer != nil {
		fw.timer.Stop()
	}
	fw.timer = time.AfterFunc(fw.debounce, func() {
		fw.mu.Lock()
		suppressed := fw.suppressed
		fw.mu.Unlock()

		if !suppressed {
			select {
			case fw.changeCh <- struct{}{}:
			default:
			}
		}
	})
}

func (fw *FolderWatcher) addRecursive(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip unreadable paths without failing entire watch
		}
		if d.IsDir() {
			name := d.Name()
			if isIgnoredDir(name) && path != dir {
				return filepath.SkipDir
			}
			_ = fw.watcher.Add(path)
		}
		return nil
	})
}

func isIgnoredDir(name string) bool {
	return strings.HasPrefix(name, ".")
}

func isIgnoredFile(name string) bool {
	// Hidden files (e.g. .git, .goutputstream, .nuvem)
	if strings.HasPrefix(name, ".") {
		return true
	}
	// Backup or temporary editor files
	if strings.HasSuffix(name, "~") || strings.HasSuffix(name, "#") {
		return true
	}
	// Common lock and temporary patterns
	lower := strings.ToLower(name)
	if strings.Contains(lower, ".~lock.") || strings.HasPrefix(lower, "~$") {
		return true
	}
	if strings.HasSuffix(lower, ".tmp") ||
		strings.HasSuffix(lower, ".swp") ||
		strings.HasSuffix(lower, ".swo") ||
		strings.HasSuffix(lower, ".bak") ||
		strings.HasSuffix(lower, ".crdownload") ||
		strings.HasSuffix(lower, ".part") {
		return true
	}
	return false
}
