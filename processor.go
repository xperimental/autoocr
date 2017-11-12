package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type processor struct {
	ctx         context.Context
	inputDir    string
	outputDir   string
	pdfSandwich string
	languages   string
	trigger     chan struct{}
}

func newProcessor(ctx context.Context, inputDir, pdfSandwich, languages, outputDir string) (*processor, error) {
	return &processor{
		ctx:         ctx,
		inputDir:    inputDir,
		outputDir:   outputDir,
		pdfSandwich: pdfSandwich,
		languages:   languages,
		trigger:     make(chan struct{}),
	}, nil
}

func (p *processor) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-p.ctx.Done():
				log.Print("Stopping processor.")
				return
			case <-p.trigger:
				if err := p.run(); err != nil {
					log.Errorf("Error during processing: %s", err)
				}
			}
		}
	}()
}

func (p *processor) Run() {
	p.trigger <- struct{}{}
}

func (p *processor) run() error {
	infos, err := ioutil.ReadDir(p.inputDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %s", err)
	}

	files := []string{}
	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		ext := filepath.Ext(info.Name())
		if ext != ".pdf" {
			continue
		}

		files = append(files, filepath.Join(p.inputDir, info.Name()))
	}

	return p.processFiles(files)
}

func (p *processor) processFiles(files []string) error {
	for _, file := range files {
		log.Printf("Processing %s ...", file)
		start := time.Now()
		if err := p.processFile(file); err != nil {
			return fmt.Errorf("error processing file %s: %s", file, err)
		}

		log.Printf("Processed %s in %s.", file, time.Since(start))
	}
	return nil
}

func (p *processor) processFile(file string) error {
	debugFile := filepath.Join(p.outputDir, filepath.Base(file)+".debug.txt")
	debug, err := os.Create(debugFile)
	if err != nil {
		return fmt.Errorf("error creating debug file: %s", err)
	}
	defer debug.Close()

	outFile := filepath.Join(p.outputDir, filepath.Base(file))
	args := []string{
		"-o", outFile,
		"-lang", p.languages,
		"-rgb",
		file,
	}

	command := exec.Command(p.pdfSandwich, args...)
	command.Stdout = debug
	command.Stderr = debug

	if err := command.Run(); err != nil {
		return fmt.Errorf("error running pdfsandwich: %s", err)
	}

	if err := os.Remove(file); err != nil {
		return fmt.Errorf("error deleting original: %s", err)
	}

	if err := os.Remove(debugFile); err != nil {
		return fmt.Errorf("error removing debug file: %s", err)
	}

	return nil
}
