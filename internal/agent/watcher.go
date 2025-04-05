package agent

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	dir     string
	re      *regexp.Regexp
	workers map[string]*FileWorker

	stop chan struct{}
}

func NewFileWatcher(path string) (*FileWatcher, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must be absolute: %s", path)
	}

	dir, pattern := filepath.Split(path)

	pattern = "^" + pattern + "$"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	return &FileWatcher{
		dir:     dir,
		re:      re,
		workers: make(map[string]*FileWorker),
		stop:    make(chan struct{}),
	}, nil
}

func (fw *FileWatcher) startWorker(path string, watcher *fsnotify.Watcher) {
	name := filepath.Base(path)
	match := fw.re.FindStringSubmatch(name)
	if match == nil {
		return
	}

	data := make(map[string]string)
	for i, name := range fw.re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		data[name] = match[i]
	}

	worker := NewFileWorker(path, data)
	fw.workers[name] = worker
	watcher.Add(path)
}

func (fw *FileWatcher) stopWorker(path string, watcher *fsnotify.Watcher) {
	name := filepath.Base(path)

	if _, ok := fw.workers[name]; ok {
		delete(fw.workers, name)
		watcher.Remove(path)
	}
}

func (fw *FileWatcher) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to start watcher", "dir", fw.dir, "error", err)
		return
	}
	defer watcher.Close()

	entries, err := os.ReadDir(fw.dir)
	if err != nil {
		slog.Error("failed to read directory", "dir", fw.dir, "error", err)
		return
	}

	watcher.Add(fw.dir)

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		name := entry.Name()
		path := filepath.Join(fw.dir, name)
		fw.startWorker(path, watcher)
	}

	for {
		select {
		case <-fw.stop:
			return
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}

			if event.Has(fsnotify.Write) {
				name := filepath.Base(event.Name)
				if worker, ok := fw.workers[name]; ok {
					worker.Handle()
				}
			} else if event.Has(fsnotify.Create) {
				stat, err := os.Stat(event.Name)
				if err != nil || !stat.Mode().IsRegular() {
					continue
				}

				fw.startWorker(event.Name, watcher)
			} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				fw.stopWorker(event.Name, watcher)
			}
		}
	}
}

// Start begins monitoring log files specified by the path pattern.
// For each matched file, it launches a goroutine that watches for changes.
func (fw *FileWatcher) Start() {
	go fw.watch()
}

// Stop stops monitoring the log files and closes the watcher.
func (fw *FileWatcher) Stop() {
	if fw.stop != nil {
		close(fw.stop)
	}
}
