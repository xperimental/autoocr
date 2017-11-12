package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/xperimental/autoocr/processor"
	"github.com/xperimental/autoocr/watcher"
)

func main() {
	config, err := parseArgs()
	if err != nil {
		log.Fatalf("Error parsing arguments: %s", err)
	}

	logger := config.CreateLogger()

	logger.Debugf("Input: %s", config.InputDir)
	logger.Debugf("Output: %s", config.OutputDir)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	watcher, err := watcher.New(ctx, logger, config.InputDir, config.Delay)
	if err != nil {
		logger.Fatalf("Error creating watcher: %s", err)
	}
	watcher.Start(wg)

	processor, err := processor.New(ctx,
		logger,
		config.InputDir,
		config.PdfSandwich,
		config.Languages,
		config.OutputDir,
		config.KeepOriginal,
		os.FileMode(config.OutPermissions))
	if err != nil {
		logger.Fatalf("Error creating processor: %s", err)
	}
	processor.Start(wg)

	go func() {
		wg.Add(1)
		defer wg.Done()

		abort := make(chan os.Signal)
		signal.Notify(abort, syscall.SIGINT)
		defer signal.Stop(abort)

		logger.Info("Waiting for changes...")
		for {
			select {
			case <-abort:
				cancel()
				return
			case <-watcher.Trigger:
				processor.Trigger()
			}
		}
	}()

	wg.Wait()
	logger.Info("All done. Exiting.")
}
