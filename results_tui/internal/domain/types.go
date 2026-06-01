package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	DefaultXAxis      = "Simulation time"
	DefaultOutputFile = "combined_results.html"
)

type ResultFile struct {
	Path           string
	DirName        string
	FileName       string
	Temperature    string
	RunLabel       string
	Headers        []string
	NumericColumns map[string]bool
	ReadError      error
}

func (f ResultFile) Valid() bool {
	return f.ReadError == nil
}

func (f ResultFile) Label() string {
	label := f.RunLabel
	if label == "" {
		label = strings.TrimSuffix(f.FileName, filepath.Ext(f.FileName))
	}
	if f.Temperature != "" {
		return fmt.Sprintf("%s %s", f.Temperature, label)
	}
	return label
}

type ColumnOption struct {
	Name    string
	Source  map[int][]string
	Preview string
}

type PlotConfig struct {
	XAxis       string
	YAxis       string
	MergeSuffix bool
}

func (p PlotConfig) Label() string {
	mode := "strict"
	if p.MergeSuffix {
		mode = "suffix"
	}
	return fmt.Sprintf("%s / %s (%s)", p.YAxis, p.XAxis, mode)
}
