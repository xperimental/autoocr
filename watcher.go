package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	ctx      context.Context
	inputDir string
	delay    time.Duration
	watcher  *fsnotify.Watcher
	timer    *time.Timer

	Trigger chan struct{}
}

func newWatcher(ctx context.Context, inputDir string, delay time.Duration) (*watcher, error) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %s", err)
	}

	if err := watch.Add(inputDir); err != nil {
		return nil, fmt.Errorf("error adding watch for input: %s", err)
	}

	return &watcher{
		ctx:      ctx,
		inputDir: inputDir,
		delay:    delay,
		watcher:  watch,
		timer:    time.NewTimer(delay),
		Trigger:  make(chan struct{}),
	}, nil
}

func (w *watcher) Close() error {
	return w.watcher.Close()
}

func (w *watcher) resetTimer() {
	if !w.timer.Stop() {
		<-w.timer.C
	}
	w.timer.Reset(w.delay)
}

func (w *watcher) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-w.ctx.Done():
				log.Info("Stopping watcher.")
				return
			case <-w.watcher.Events:
				w.timer.Reset(w.delay)
			case <-w.timer.C:
				w.Trigger <- struct{}{}
			}
		}
	}()
}
