package main 

import (
	"fmt"
	"strings"
)

// Column represents a single column in the table
type Column struct {
	Header    string
	MinWidth  int
	Extractor func(interface{}) string
}

// FormatTable creates a formatted table from any slice of data
func FormatTable(data []interface{}, columns []Column) string {
	var sb strings.Builder

	if len(data) == 0 {
		return ""
	}

	// Calculate max widths for each column
	maxWidths := make([]int, len(columns))
	for i, col := range columns {
		maxWidths[i] = len(col.Header)
		if col.MinWidth > maxWidths[i] {
			maxWidths[i] = col.MinWidth
		}
	}

	// Extract all rows and update max widths
	rows := make([][]string, len(data))
	for i, item := range data {
		rows[i] = make([]string, len(columns))
		for j, col := range columns {
			rows[i][j] = col.Extractor(item)
			if len(rows[i][j]) > maxWidths[j] {
				maxWidths[j] = len(rows[i][j])
			}
		}
	}

	// Print header
	headerParts := make([]string, len(columns))
	for i, col := range columns {
		headerParts[i] = fmt.Sprintf("%-*s", maxWidths[i], col.Header)
	}
	fmt.Fprintf(&sb, "%s\n", strings.Join(headerParts, "\t"))

	// Print rows
	for _, row := range rows {
		rowParts := make([]string, len(columns))
		for i, cell := range row {
			rowParts[i] = fmt.Sprintf("%-*s", maxWidths[i], cell)
		}
		fmt.Fprintf(&sb, "%s\n", strings.Join(rowParts, "\t"))
	}

	return sb.String()
}
