package bulk

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseFileCSV(t *testing.T) {
	filePath := writeTempFile(t, "rows.csv", "date,amount,comment\n2026-04-10,1200,lunch\n2026-04-11,3400,book\n")

	got, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	want := []Row{
		{
			Index: 1,
			Fields: map[string]string{
				"date":    "2026-04-10",
				"amount":  "1200",
				"comment": "lunch",
			},
		},
		{
			Index: 2,
			Fields: map[string]string{
				"date":    "2026-04-11",
				"amount":  "3400",
				"comment": "book",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseFile() = %#v, want %#v", got, want)
	}
}

func TestParseFileJSON(t *testing.T) {
	filePath := writeTempFile(t, "rows.json", `[
		{"date":"2026-04-10","amount":1200,"comment":"lunch"},
		{"date":"2026-04-11","amount":3400,"comment":"book"}
	]`)

	got, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	want := []Row{
		{
			Index: 1,
			Fields: map[string]string{
				"date":    "2026-04-10",
				"amount":  "1200",
				"comment": "lunch",
			},
		},
		{
			Index: 2,
			Fields: map[string]string{
				"date":    "2026-04-11",
				"amount":  "3400",
				"comment": "book",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseFile() = %#v, want %#v", got, want)
	}
}

func TestParseFileEmptyFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{
			name:     "csv",
			fileName: "empty.csv",
		},
		{
			name:     "json",
			fileName: "empty.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := writeTempFile(t, tt.fileName, "")

			_, err := ParseFile(filePath)
			if err == nil {
				t.Fatal("ParseFile() error = nil, want error")
			}
		})
	}
}

func TestParseFileUnsupportedExtension(t *testing.T) {
	filePath := writeTempFile(t, "rows.txt", "date,amount,comment\n2026-04-10,1200,lunch\n")

	_, err := ParseFile(filePath)
	if err == nil {
		t.Fatal("ParseFile() error = nil, want error")
	}
}

func TestParseFileCSVEmptyField(t *testing.T) {
	filePath := writeTempFile(t, "rows.csv", "date,amount,comment\n2026-04-10,,lunch\n")

	got, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	want := []Row{
		{
			Index: 1,
			Fields: map[string]string{
				"date":    "2026-04-10",
				"amount":  "",
				"comment": "lunch",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseFile() = %#v, want %#v", got, want)
	}
}

func writeTempFile(t *testing.T, fileName string, content string) string {
	t.Helper()

	dir := t.TempDir()
	filePath := filepath.Join(dir, fileName)

	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return filePath
}
