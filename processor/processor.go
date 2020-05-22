package processor

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Config contains the configuration options for Processor.
type Config struct {
	InputDir          string
	OutputDir         string
	OutputPermissions os.FileMode
	PdfSandwichPath   string
	Languages         string
	KeepOriginal      bool
}

// Processor processes PDF files.
type Processor struct {
	ctx     context.Context
	log     *logrus.Entry
	cfg     Config
	trigger chan struct{}
}

// New creates a new processor that will process all PDF files in a directory.
// Call Start to actually start waiting for signals.
func New(ctx context.Context, logger *logrus.Logger, cfg Config) (*Processor, error) {
	return &Processor{
		ctx:     ctx,
		log:     logger.WithField("component", "processor"),
		cfg:     cfg,
		trigger: make(chan struct{}),
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
				p.log.Print("Stopping.")
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
	infos, err := ioutil.ReadDir(p.cfg.InputDir)
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

		files = append(files, filepath.Join(p.cfg.InputDir, info.Name()))
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
	statusFile := filepath.Join(p.cfg.InputDir, filepath.Base(file)+".processing")
	status, err := os.OpenFile(statusFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error creating status file: %s", err)
	}
	status.Close()
	defer os.Remove(statusFile)

	debugFile := filepath.Join(p.cfg.OutputDir, filepath.Base(file)+".debug.txt")
	debug, err := os.OpenFile(debugFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, p.cfg.OutputPermissions)
	if err != nil {
		return fmt.Errorf("error creating debug file: %s", err)
	}
	defer debug.Close()

	outFile := filepath.Join(p.cfg.OutputDir, filepath.Base(file))
	args := []string{
		"-o", outFile,
		"-lang", p.cfg.Languages,
		"-rgb",
		file,
	}

	command := exec.Command(p.cfg.PdfSandwichPath, args...)
	command.Stdout = debug
	command.Stderr = debug

	if err := command.Run(); err != nil {
		return fmt.Errorf("error running pdfsandwich: %s", err)
	}

	if err := os.Chmod(outFile, p.cfg.OutputPermissions); err != nil {
		return fmt.Errorf("error setting permissions: %s", err)
	}

	if p.cfg.KeepOriginal {
		backupFile := filepath.Join(p.cfg.OutputDir, filepath.Base(file)+".backup")
		if err := copyFile(file, backupFile, p.cfg.OutputPermissions); err != nil {
			return fmt.Errorf("error creating backup: %s", err)
		}
	}

	if err := os.Remove(file); err != nil {
		return fmt.Errorf("error deleting original: %s", err)
	}

	if err := os.Remove(debugFile); err != nil {
		return fmt.Errorf("error removing debug file: %s", err)
	}

	return nil
}

func copyFile(src, dst string, perms os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source: %s", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return fmt.Errorf("error creating target: %s", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copying: %s", err)
	}

	return nil
}
