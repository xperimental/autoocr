package processor

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

// Processor processes PDF files.
type Processor struct {
	ctx         context.Context
	log         *logrus.Entry
	inputDir    string
	outputDir   string
	pdfSandwich string
	languages   string
	trigger     chan struct{}
}

// New creates a new processor that will process all PDF files in a directory.
// Call Start to actually start waiting for signals.
func New(ctx context.Context, logger *logrus.Logger, inputDir, pdfSandwich, languages, outputDir string) (*Processor, error) {
	return &Processor{
		ctx:         ctx,
		log:         logger.WithField("component", "processor"),
		inputDir:    inputDir,
		outputDir:   outputDir,
		pdfSandwich: pdfSandwich,
		languages:   languages,
		trigger:     make(chan struct{}),
	}, nil
}

// Start starts the background routine. The waitgroup is used for tracking the shutdown.
func (p *Processor) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-p.ctx.Done():
				p.log.Print("Stopping processor.")
				return
			case <-p.trigger:
				if err := p.run(); err != nil {
					p.log.Errorf("Error during processing: %s", err)
				}
			}
		}
	}()
}

// Trigger triggers one processing run.
func (p *Processor) Trigger() {
	p.trigger <- struct{}{}
}

func (p *Processor) run() error {
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

func (p *Processor) processFiles(files []string) error {
	for _, file := range files {
		filelog := p.log.WithField("file", file)
		filelog.Print("Start processing.")
		start := time.Now()
		if err := p.processFile(file); err != nil {
			filelog.WithError(err).Error("Error processing file.")
			continue
		}

		filelog.Printf("Processing successful in %s.", time.Since(start))
	}
	return nil
}

func (p *Processor) processFile(file string) error {
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
