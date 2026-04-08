// Package formatter provides output formatting utilities for CLI commands.
package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
)

const (
	FormatJSON  = "json"
	FormatTable = "table"
)

var Styles = struct {
	Title   lipgloss.Style
	Header  lipgloss.Style
	Key     lipgloss.Style
	Value   lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Subtle  lipgloss.Style
}{
	Title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
	Header:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")),
	Key:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")),
	Value:   lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
	Success: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")),
	Error:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")),
	Warning: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
	Subtle:  lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
}

// OutputJSON writes formatted JSON to the provided writer.
func OutputJSON(w io.Writer, v any) error {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	if _, err := fmt.Fprintln(w, string(output)); err != nil {
		return fmt.Errorf("json write failed: %w", err)
	}

	return nil
}

// RenderTable renders a table with the provided header and rows.
func RenderTable(w io.Writer, header []string, rows [][]string) {
	table := tablewriter.NewWriter(w)

	headerAny := make([]any, len(header))
	for i, h := range header {
		headerAny[i] = h
	}
	table.Header(headerAny...)

	_ = table.Bulk(convertToAny(rows))
	_ = table.Render()
}

// FormatKeyValue formats a key-value pair with optional color styling.
func FormatKeyValue(key, value string) string {
	if !IsColorEnabled() {
		return fmt.Sprintf("%s: %s", key, value)
	}

	return fmt.Sprintf("%s %s", Styles.Key.Render(key+":"), Styles.Value.Render(value))
}

// TruncateString truncates a string to max characters and appends ellipsis.
func TruncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max]) + "..."
}

// IsColorEnabled reports whether colored output should be enabled.
func IsColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func convertToAny(rows [][]string) [][]any {
	result := make([][]any, len(rows))
	for i, row := range rows {
		result[i] = make([]any, len(row))
		for j, cell := range row {
			result[i][j] = cell
		}
	}
	return result
}
