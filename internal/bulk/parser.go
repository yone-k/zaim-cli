package bulk

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ParseFile parses a bulk input file into rows.
func ParseFile(filePath string) ([]Row, error) {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".csv":
		return parseCSVFile(filePath)
	case ".json":
		return parseJSONFile(filePath)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", filepath.Ext(filePath))
	}
}

func parseCSVFile(filePath string) ([]Row, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("empty file")
		}
		return nil, err
	}

	var rows []Row
	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		fields := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(record) {
				fields[header] = record[i]
			} else {
				fields[header] = ""
			}
		}

		rows = append(rows, Row{
			Index:  len(rows) + 1,
			Fields: fields,
		})
	}

	return rows, nil
}

func parseJSONFile(filePath string) ([]Row, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var values []map[string]any
	if err := json.NewDecoder(file).Decode(&values); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("empty file")
		}
		return nil, err
	}

	rows := make([]Row, 0, len(values))
	for i, value := range values {
		fields := make(map[string]string, len(value))
		for key, field := range value {
			switch v := field.(type) {
			case float64:
				if v == float64(int64(v)) {
					fields[key] = fmt.Sprintf("%d", int64(v))
				} else {
					fields[key] = fmt.Sprintf("%v", v)
				}
			default:
				fields[key] = fmt.Sprintf("%v", v)
			}
		}

		rows = append(rows, Row{
			Index:  i + 1,
			Fields: fields,
		})
	}

	return rows, nil
}
