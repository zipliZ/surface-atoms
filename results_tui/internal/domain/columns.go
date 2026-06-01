package domain

import (
	"sort"
	"strings"
)

func ComputeColumns(files []ResultFile, mergeSuffix bool) []ColumnOption {
	if len(files) == 0 {
		return nil
	}

	var intersection map[string]map[int][]string
	for fileIndex, file := range files {
		perFile := make(map[string][]string)
		for _, header := range file.Headers {
			if !file.NumericColumns[header] {
				continue
			}
			name := header
			if mergeSuffix {
				name = NormalizeSuffix(header)
			}
			perFile[name] = append(perFile[name], header)
		}

		if intersection == nil {
			intersection = make(map[string]map[int][]string)
			for name, headers := range perFile {
				intersection[name] = map[int][]string{fileIndex: headers}
			}
			continue
		}

		for name := range intersection {
			headers, ok := perFile[name]
			if !ok {
				delete(intersection, name)
				continue
			}
			intersection[name][fileIndex] = headers
		}
	}

	columns := make([]ColumnOption, 0, len(intersection))
	for name, source := range intersection {
		columns = append(columns, ColumnOption{
			Name:    name,
			Source:  source,
			Preview: previewSource(source),
		})
	}
	sort.Slice(columns, func(i, j int) bool {
		if columns[i].Name == DefaultXAxis {
			return true
		}
		if columns[j].Name == DefaultXAxis {
			return false
		}
		return columns[i].Name < columns[j].Name
	})

	return columns
}

func NormalizeSuffix(header string) string {
	parts := strings.Split(header, " - ")
	if len(parts) < 2 {
		return header
	}
	return strings.TrimSpace(parts[len(parts)-1])
}

func previewSource(source map[int][]string) string {
	var examples []string
	for _, headers := range source {
		examples = append(examples, headers...)
		if len(examples) >= 2 {
			break
		}
	}
	sort.Strings(examples)
	if len(examples) == 0 {
		return ""
	}
	if len(examples) > 2 {
		examples = examples[:2]
	}
	return strings.Join(examples, ", ")
}
