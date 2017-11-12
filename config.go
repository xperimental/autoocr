package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
)

type logFormat string

func (f *logFormat) String() string {
	return string(*f)
}

func (f *logFormat) Set(value string) error {
	switch value {
	case "plain":
		*f = logFormatPlain
	case "json":
		*f = logFormatJSON
	default:
		return fmt.Errorf("invalid log format: %s", value)
	}
	return nil
}

func (f *logFormat) Type() string {
	return "string"
}

const (
	logFormatPlain logFormat = "plain"
	logFormatJSON  logFormat = "json"
)

type logLevel logrus.Level

func (f *logLevel) String() string {
	return logrus.Level(*f).String()
}

func (f *logLevel) Set(value string) error {
	lvl, err := logrus.ParseLevel(value)
	if err != nil {
		return err
	}

	*f = logLevel(lvl)
	return nil
}

func (f *logLevel) Type() string {
	return "string"
}

type fileMode os.FileMode

func (f *fileMode) String() string {
	return logrus.Level(*f).String()
}

func (f *fileMode) Set(value string) error {
	val, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	*f = fileMode(val)
	return nil
}

func (f *fileMode) Type() string {
	return "uint32"
}

type config struct {
	InputDir       string
	OutputDir      string
	PdfSandwich    string
	Languages      string
	Delay          time.Duration
	LogFormat      logFormat
	LogLevel       logLevel
	KeepOriginal   bool
	OutPermissions fileMode
}

func (c config) CreateLogger() *logrus.Logger {
	var formatter logrus.Formatter = &logrus.TextFormatter{}
	if c.LogFormat == logFormatJSON {
		formatter = &logrus.JSONFormatter{}
	}
	return &logrus.Logger{
		Out:       os.Stdout,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.Level(c.LogLevel),
	}
}

func parseArgs() (config, error) {
	cfg := config{
		InputDir:       "input",
		OutputDir:      "output",
		PdfSandwich:    "pdfsandwich",
		Languages:      "deu+eng",
		Delay:          5 * time.Second,
		LogFormat:      logFormatPlain,
		LogLevel:       logLevel(logrus.InfoLevel),
		KeepOriginal:   true,
		OutPermissions: fileMode(0644),
	}
	pflag.StringVarP(&cfg.InputDir, "input", "i", cfg.InputDir, "Directory to use for input.")
	pflag.StringVarP(&cfg.OutputDir, "output", "o", cfg.OutputDir, "Directory to use for output.")
	pflag.StringVar(&cfg.PdfSandwich, "pdf-sandwich", cfg.PdfSandwich, "Path to pdfsandwich utility.")
	pflag.StringVar(&cfg.Languages, "languages", cfg.Languages, "OCR Languages to use.")
	pflag.DurationVar(&cfg.Delay, "delay", cfg.Delay, "Processing delay after receiving watch events.")
	pflag.Var(&cfg.LogFormat, "log-format", "Logging format to use.")
	pflag.Var(&cfg.LogLevel, "log-level", "Logging level to show.")
	pflag.BoolVar(&cfg.KeepOriginal, "keep-original", cfg.KeepOriginal, "Keep backup of original file.")
	pflag.Var(&cfg.OutPermissions, "permissions", "Permissions on output files.")
	pflag.Parse()

	if len(cfg.InputDir) == 0 {
		return cfg, errors.New("input directory can not be empty")
	}

	if len(cfg.OutputDir) == 0 {
		return cfg, errors.New("output directory can not be empty")
	}

	if len(cfg.PdfSandwich) == 0 {
		return cfg, errors.New("path to pdfsandwich can not be empty")
	}

	if len(cfg.Languages) == 0 {
		return cfg, errors.New("languages can not be empty")
	}

	if cfg.Delay <= 0 {
		return cfg, errors.New("delay can not be smaller or equal to zero")
	}

	return cfg, nil
}
