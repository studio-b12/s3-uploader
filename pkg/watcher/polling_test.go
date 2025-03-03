package watcher

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(levelStr)); err != nil {
			panic(fmt.Sprintf("failed to parse LOG_LEVEL: %v", err))
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
	}
}

func TestPollingWatcher(t *testing.T) {
	mfs := fstest.MapFS{
		"foo.txt":     {Data: []byte("foo"), ModTime: time.Time{}},
		"dir/bar.txt": {Data: []byte("bar"), ModTime: time.Time{}},
	}

	pw := NewPollingWatcher(mfs, time.Second, 10)

	t.Run("empty after first poll", func(t *testing.T) {
		assert.Nil(t, pw.poll())
		assert.Zero(t, len(pw.Results()))
	})

	t.Run("initial index", func(t *testing.T) {
		assert.Nil(t, pw.poll())
		results := pull(pw.Results())
		assert.Contains(t, results, File{Path: "foo.txt"})
		assert.Contains(t, results, File{Path: "dir/bar.txt"})
	})

	t.Run("modified mtime", func(t *testing.T) {
		mfs["foo.txt"] = &fstest.MapFile{Data: []byte("foo"), ModTime: time.Time{}.Add(time.Second)}

		assert.Nil(t, pw.poll())
		assert.Zero(t, len(pw.Results()))

		assert.Nil(t, pw.poll())
		results := pull(pw.Results())
		assert.Contains(t, results, File{Path: "foo.txt"})
	})

	t.Run("modified size", func(t *testing.T) {
		mfs["dir/bar.txt"] = &fstest.MapFile{Data: []byte("bar bar"), ModTime: time.Time{}}

		assert.Nil(t, pw.poll())
		assert.Zero(t, len(pw.Results()))

		assert.Nil(t, pw.poll())
		results := pull(pw.Results())
		assert.Contains(t, results, File{Path: "dir/bar.txt"})
	})

	t.Run("modified between polls", func(t *testing.T) {
		for i := range 3 {
			mfs["foo.txt"] = &fstest.MapFile{Data: []byte("foo"), ModTime: time.Time{}.Add(time.Duration(i+1) * time.Second)}

			assert.Nil(t, pw.poll())
			assert.Zero(t, len(pw.Results()))
		}

		assert.Nil(t, pw.poll())
		results := pull(pw.Results())
		assert.Contains(t, results, File{Path: "foo.txt"})
	})

	t.Run("new file", func(t *testing.T) {
		mfs["bar/baz.txt"] = &fstest.MapFile{Data: []byte("baz"), ModTime: time.Time{}}

		assert.Nil(t, pw.poll())
		assert.Zero(t, len(pw.Results()))

		assert.Nil(t, pw.poll())
		results := pull(pw.Results())
		assert.Contains(t, results, File{Path: "bar/baz.txt"})
	})
}

func pull[T any](c <-chan T) (vals []T) {
	for range len(c) {
		vals = append(vals, <-c)
	}
	return vals
}
