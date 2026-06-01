package results

import (
	"errors"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"results_tui/internal/domain"
)

func readExcelMetadata(path string) ([]string, map[string]bool, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	rows, err := file.GetRows("Sheet1")
	if err != nil {
		return nil, nil, err
	}
	if len(rows) < 2 || len(rows[0]) == 0 {
		return nil, nil, errors.New("Sheet1 has no data")
	}

	headers := make([]string, 0, len(rows[0]))
	seen := make(map[string]bool)
	for _, header := range rows[0] {
		header = strings.TrimSpace(header)
		if header == "" || seen[header] {
			continue
		}
		headers = append(headers, header)
		seen[header] = true
	}

	numeric := make(map[string]bool)
	for colIndex, header := range rows[0] {
		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}
		for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
			row := rows[rowIndex]
			if colIndex >= len(row) {
				continue
			}
			value := strings.TrimSpace(row[colIndex])
			if value == "" {
				continue
			}
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				numeric[header] = true
				break
			}
		}
	}

	return headers, numeric, metadataError(headers, numeric)
}

func ReadExcelColumns(path string, required []string, mergeSuffix bool) (map[string][]float64, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rows, err := file.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, errors.New("no data")
	}

	requiredSet := make(map[string]bool)
	for _, name := range required {
		requiredSet[name] = true
	}

	indexes := make(map[int]string)
	for i, header := range rows[0] {
		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}
		if requiredSet[header] || (mergeSuffix && requiredSet[domain.NormalizeSuffix(header)]) {
			indexes[i] = header
		}
	}
	if len(indexes) == 0 {
		return nil, errors.New("required columns not found")
	}

	values := make(map[string][]float64)
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		for index, name := range indexes {
			if index >= len(row) {
				continue
			}
			value, err := strconv.ParseFloat(strings.TrimSpace(row[index]), 64)
			if err != nil {
				continue
			}
			values[name] = append(values[name], value)
		}
	}
	return values, nil
}

func FirstColumn(columns map[string][]float64, logical string, mergeSuffix bool) []float64 {
	if values, ok := columns[logical]; ok {
		return values
	}
	if !mergeSuffix {
		return nil
	}

	var names []string
	for name := range columns {
		if domain.NormalizeSuffix(name) == logical {
			names = append(names, name)
		}
	}
	sortStrings(names)
	if len(names) == 0 {
		return nil
	}
	return columns[names[0]]
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
