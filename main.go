package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/xperimental/autoocr/processor"
	"github.com/xperimental/autoocr/watcher"
)

var log = logrus.New()

func main() {
	config, err := parseArgs()
	if err != nil {
		log.Fatalf("Error parsing arguments: %s", err)
	}

	log.Printf("Input:  %s", config.InputDir)
	log.Printf("Output: %s", config.OutputDir)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	watcher, err := watcher.New(ctx, log, config.InputDir, config.Delay)
	if err != nil {
		log.Fatalf("Error creating watcher: %s", err)
	}
	watcher.Start(wg)

	processor, err := processor.New(ctx, log, config.InputDir, config.PdfSandwich, config.Languages, config.OutputDir)
	if err != nil {
		log.Fatalf("Error creating processor: %s", err)
	}
	processor.Start(wg)

	go func() {
		wg.Add(1)
		defer wg.Done()

		abort := make(chan os.Signal)
		signal.Notify(abort, syscall.SIGINT)
		defer signal.Stop(abort)

		log.Println("Waiting for changes...")
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
	log.Println("All done. Exiting.")
}
