package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher *fsnotify.Watcher
	mu      sync.Mutex
	paths   []string
}

func New() (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{watcher: w}, nil
}

func (w *FileWatcher) AddPath(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paths = append(w.paths, path)
	return w.watcher.Add(path)
}

func (w *FileWatcher) Watch(ctx context.Context, callback func(path string)) error {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			// Only watch for modification and creation (ignoring temporary files)
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if w.isCodeFile(event.Name) {
					callback(event.Name)
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("watcher error: %v\n", err)
		case <-ctx.Done():
			return w.watcher.Close()
		}
	}
}

func (w *FileWatcher) isCodeFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".gd", ".cpp", ".h", ".cs", ".md", ".xml", ".tscn":
		return true
	}
	return false
}
