package configdefaults

import (
	"os"

	"gopkg.in/yaml.v3"

	"results_tui/internal/domain"
)

type fileConfig struct {
	Simulating simulatingConfig `yaml:"simulating"`
}

type simulatingConfig struct {
	GraphicsToPlot []graphicToPlot `yaml:"graphicsToPlot"`
}

type graphicToPlot struct {
	XAxis string `yaml:"xAxis"`
	YAxis string `yaml:"yAxis"`
}

func Load(path string, available []domain.ColumnOption) ([]domain.PlotConfig, bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var cfg fileConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, false
	}

	availableNames := make(map[string]bool, len(available))
	for _, col := range available {
		availableNames[col.Name] = true
	}

	graphs := make([]domain.PlotConfig, 0, len(cfg.Simulating.GraphicsToPlot))
	seen := make(map[domain.PlotConfig]bool)
	for _, item := range cfg.Simulating.GraphicsToPlot {
		if item.XAxis == "" || item.YAxis == "" {
			continue
		}
		if !availableNames[item.XAxis] || !availableNames[item.YAxis] {
			continue
		}
		graph := domain.PlotConfig{XAxis: item.XAxis, YAxis: item.YAxis}
		if seen[graph] {
			continue
		}
		graphs = append(graphs, graph)
		seen[graph] = true
	}

	return graphs, true
}
