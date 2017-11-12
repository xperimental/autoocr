package watcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

// A Watcher watches a directory for changes and waits a delay until triggering an output event.
type Watcher struct {
	ctx      context.Context
	log      *logrus.Entry
	inputDir string
	delay    time.Duration
	watcher  *fsnotify.Watcher
	timer    *time.Timer

	Trigger chan struct{}
}

// New creates a new filesystem watcher. Need to call Start() to actually start watching.
func New(ctx context.Context, logger *logrus.Logger, inputDir string, delay time.Duration) (*Watcher, error) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %s", err)
	}

	if err := watch.Add(inputDir); err != nil {
		return nil, fmt.Errorf("error adding watch for input: %s", err)
	}

	return &Watcher{
		ctx:      ctx,
		log:      logger.WithField("component", "watcher"),
		inputDir: inputDir,
		delay:    delay,
		watcher:  watch,
		timer:    time.NewTimer(delay),
		Trigger:  make(chan struct{}),
	}, nil
}

func (w *Watcher) resetTimer() {
	if !w.timer.Stop() {
		<-w.timer.C
	}
	w.timer.Reset(w.delay)
}

// Start starts the background routine which watches for filesystem events.
// The waitgroup is used for tracking the shutdown.
func (w *Watcher) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-w.ctx.Done():
				w.log.Info("Stopping.")
				if err := w.watcher.Close(); err != nil {
					w.log.Errorf("Error closing watcher: %s", err)
				}
				return
			case <-w.watcher.Events:
				w.timer.Reset(w.delay)
			case <-w.timer.C:
				w.Trigger <- struct{}{}
			}
		}
	}()
}
