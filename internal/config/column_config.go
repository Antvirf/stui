package config

import (
	"errors"
	"strconv"
	"strings"
)

type ColumnConfig struct {
	Name            string
	Width           int
	DividedByColumn bool
}

func GetColumnNames(columnConfigs *[]ColumnConfig) (columns []string) {
	for _, col := range *columnConfigs {
		columns = append(columns, col.Name)
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

		if strings.Contains(part, ":") {
			subParts := strings.Split(part, ":")
			if len(subParts) != 2 {
				return nil, errors.New("invalid column format: " + part)
			}
			col.Name = subParts[0]
			width, err := strconv.Atoi(subParts[1])
			if err != nil {
				return nil, errors.New("invalid width value: " + subParts[1])
			}
			col.Width = width
		} else {
			col.Name = part
			col.Width = DefaultColumnWidth
		}

		// Check if column contains '/', in which case it is a DividedByColumn
		if strings.Contains(part, "//") {
			col.DividedByColumn = true
		}
		configs = append(configs, col)
	}

	return &configs, nil
}
