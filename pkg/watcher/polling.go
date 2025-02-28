package watcher

import (
	"io/fs"
	"log/slog"
	"sync"
	"time"
)

type File struct {
	Path string
}

type PollingWatcher struct {
	baseFs       fs.FS
	interval     time.Duration
	fileTracking *sync.Map
	stopChan     chan struct{}
	resultChan   chan File
}

func NewPollingWatcher(baseFs fs.FS, interval time.Duration, resultBufferSize int) *PollingWatcher {
	return &PollingWatcher{
		baseFs:       baseFs,
		interval:     interval,
		fileTracking: &sync.Map{},
		stopChan:     make(chan struct{}),
		resultChan:   make(chan File, resultBufferSize),
	}
}

func (t *PollingWatcher) Start() {
	ticker := time.NewTicker(t.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					if err := t.poll(); err != nil {
						slog.Error("file walk failed", "err", err)
					}
				}()
			case <-t.stopChan:
				return
			}
		}
	}()
}

func (t *PollingWatcher) Stop() {
	close(t.resultChan)
	t.stopChan <- struct{}{}
}

func (t *PollingWatcher) Results() <-chan File {
	return t.resultChan
}

type fileInfo struct {
	fs.FileInfo
	changed bool
}

func (t *PollingWatcher) poll() error {
	err := fs.WalkDir(t.baseFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			slog.Error("failed stat file", "path", path, "err", err)
			return nil
		}

		slog.Debug("retrieved file info",
			"path", path, "mtime", info.ModTime(), "size", info.Size())

		trackedInfo, ok := t.getTrackedInfo(path)
		if !ok {
			t.fileTracking.Store(path, &fileInfo{FileInfo: info, changed: true})
			return nil
		}

		slog.Debug("tracked file info",
			"path", path, "mtime", trackedInfo.ModTime(), "size", trackedInfo.Size())

		fi := &fileInfo{FileInfo: info}

		// NOTE: Tracking just ModTIme and Size might not be accurate, because some file systems
		// do not provide a ModTime and if the sitze doesnt change but the content does, no
		// tracking event will be triggered even if the file has changed.
		// TODO: Check if file system supports mod time and if not, use file hashing.
		if info.ModTime() == trackedInfo.ModTime() && info.Size() == trackedInfo.Size() {
			if trackedInfo.changed {
				t.resultChan <- File{Path: path}
			}
		} else {
			fi.changed = true
		}

		t.fileTracking.Store(path, fi)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (t *PollingWatcher) getTrackedInfo(path string) (*fileInfo, bool) {
	info, ok := t.fileTracking.Load(path)
	if !ok {
		return nil, false
	}

	// This will panic when t.fileTracking contains something else than
	// *fileInfo, which is desired to recognize bugs early.
	return info.(*fileInfo), true
}
