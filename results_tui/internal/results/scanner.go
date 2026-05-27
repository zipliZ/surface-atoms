package results

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"results_tui/internal/domain"
)

var tempPattern = regexp.MustCompile(`T\d+K`)

func FindRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		if hasResultDirs(current) {
			return current, nil
		}
		if _, err := os.Stat(filepath.Join(current, "config.yaml")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return filepath.Abs(start)
}

func Scan(root string) []domain.ResultFile {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	var files []domain.ResultFile
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "result") {
			continue
		}

		dirPath := filepath.Join(root, entry.Name())
		children, err := os.ReadDir(dirPath)
		if err != nil {
			files = append(files, domain.ResultFile{
				Path:      dirPath,
				DirName:   entry.Name(),
				FileName:  entry.Name(),
				ReadError: err,
			})
			continue
		}

		for _, child := range children {
			if child.IsDir() || !strings.EqualFold(filepath.Ext(child.Name()), ".xlsx") {
				continue
			}

			path := filepath.Join(dirPath, child.Name())
			result := domain.ResultFile{
				Path:        path,
				DirName:     entry.Name(),
				FileName:    child.Name(),
				Temperature: parseTemperature(entry.Name(), child.Name()),
				RunLabel:    parseRunLabel(entry.Name(), child.Name()),
			}

			headers, numeric, readErr := readExcelMetadata(path)
			result.Headers = headers
			result.NumericColumns = numeric
			result.ReadError = readErr
			files = append(files, result)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files
}

func hasResultDirs(root string) bool {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "result") {
			return true
		}
	}
	return false
}

func parseTemperature(parts ...string) string {
	for _, part := range parts {
		if match := tempPattern.FindString(part); match != "" {
			return match
		}
	}
	return ""
}

func parseRunLabel(dirName, fileName string) string {
	label := strings.TrimSpace(strings.TrimPrefix(dirName, "result"))
	if temp := parseTemperature(label); temp != "" {
		label = strings.TrimSpace(strings.ReplaceAll(label, temp, ""))
	}
	if label != "" {
		return label
	}

	label = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	label = strings.TrimPrefix(label, "result_")
	if temp := parseTemperature(label); temp != "" {
		label = strings.TrimSpace(strings.ReplaceAll(label, "_"+temp, ""))
		label = strings.TrimSpace(strings.ReplaceAll(label, temp, ""))
	}
	return strings.ReplaceAll(label, "_", " ")
}

func metadataError(headers []string, numeric map[string]bool) error {
	if len(headers) == 0 {
		return errors.New("no headers found")
	}
	if len(numeric) == 0 {
		return errors.New("no numeric columns found")
	}
	return nil
}
