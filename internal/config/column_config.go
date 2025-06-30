package config

import (
	"errors"
	"log"
	"strings"
)

type ColumnConfig struct {
	RawName         string
	DisplayName     string
	Width           int
	DividedByColumn bool
	FullWidthColumn bool
}

// Get column fields returns the full expanded list of field names that
// can be used in the --format argument of sacct
func GetColumnFields(columnConfigs *[]ColumnConfig) (columns []string) {
	for _, col := range *columnConfigs {
		// Clean up configuration characters
		col.RawName = strings.ReplaceAll(col.RawName, "++", "")

		// Split apart divided columns
		if col.DividedByColumn {
			columns = append(columns, strings.Split(col.RawName, "//")...)
		} else {
			columns = append(columns, col.RawName)
		}
	}
	return
}

func parseColumnConfigLine(input string) (*[]ColumnConfig, error) {
	if input == "" {
		return nil, errors.New("cannot parse empty column config")
	}

	parts := strings.Split(input, ",")
	var configs []ColumnConfig

	for _, part := range parts {
		col := ColumnConfig{DividedByColumn: false, FullWidthColumn: false}
		col.RawName = strings.TrimSpace(part)
		col.DisplayName = strings.TrimSpace(part)

		if strings.Contains(part, "//") {
			col.DisplayName = strings.ReplaceAll(col.DisplayName, "//", "/")
			col.DividedByColumn = true
		}

		if strings.Contains(part, "++") {
			col.DisplayName = strings.ReplaceAll(col.DisplayName, "++", "")
			col.FullWidthColumn = true
		}

		configs = append(configs, col)
	}

	return &configs, nil
}

// GetColumnIndex returns the index of the column for a column config object. Panics if not found.
func GetColumnIndexFromColumnConfig(columnConfigs *[]ColumnConfig, name string) int {
	for i, col := range *columnConfigs {
		if col.RawName == name {
			return i
		}
	}
	log.Fatalf("Column %s not found in column configs %v", name, columnConfigs)
	return -1
}
