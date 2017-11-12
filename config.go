package main

import (
	"errors"
	"time"

	"github.com/spf13/pflag"
)

type config struct {
	InputDir    string
	OutputDir   string
	PdfSandwich string
	Languages   string
	Delay       time.Duration
}

func parseArgs() (config, error) {
	cfg := config{
		InputDir:    "input",
		OutputDir:   "output",
		PdfSandwich: "pdfsandwich",
		Languages:   "deu+eng",
		Delay:       5 * time.Second,
	}
	pflag.StringVarP(&cfg.InputDir, "input", "i", cfg.InputDir, "Directory to use for input.")
	pflag.StringVarP(&cfg.OutputDir, "output", "o", cfg.OutputDir, "Directory to use for output.")
	pflag.StringVar(&cfg.PdfSandwich, "pdf-sandwich", cfg.PdfSandwich, "Path to pdfsandwich utility.")
	pflag.StringVar(&cfg.Languages, "languages", cfg.Languages, "OCR Languages to use.")
	pflag.DurationVar(&cfg.Delay, "delay", cfg.Delay, "Processing delay after receiving watch events.")
	pflag.Parse()

	if len(cfg.InputDir) == 0 {
		return cfg, errors.New("input directory can not be empty")
	}

	if len(cfg.OutputDir) == 0 {
		return cfg, errors.New("utput directory can not be empty")
	}

	return cfg, nil
}
