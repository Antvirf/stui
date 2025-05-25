package config

import (
	"errors"
	"log"
	"strings"
)

type ColumnConfig struct {
	Name            string
	DividedByColumn bool
}

func GetColumnNames(columnConfigs *[]ColumnConfig) (columns []string) {
	for _, col := range *columnConfigs {
		columns = append(columns, col.Name)
	}
	return
}

// Get column fields returns the full expanded list of field names that
// can be used in the --format argument of sacct
func GetColumnFields(columnConfigs *[]ColumnConfig) (columns []string) {
	for _, col := range *columnConfigs {
		if col.DividedByColumn {
			columns = append(columns, strings.Split(col.Name, "//")...)
		} else {
			columns = append(columns, col.Name)
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
		col := ColumnConfig{DividedByColumn: false}
		col.Name = strings.TrimSpace(part)

		// Check if column contains '//', in which case it is a DividedByColumn
		if strings.Contains(part, "//") {
			col.DividedByColumn = true
		}
		configs = append(configs, col)
	}

	return &configs, nil
}

// GetColumnIndex returns the index of the column for a column config object. Panics if not found.
func GetColumnIndexFromColumnConfig(columnConfigs *[]ColumnConfig, name string) int {
	for i, col := range *columnConfigs {
		if col.Name == name {
			return i
		}
	}
	log.Fatalf("Column %s not found in column configs %v", name, columnConfigs)
	return -1
}
