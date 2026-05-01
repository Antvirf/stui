package model

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/antvirf/stui/internal/config"
)

// buildTableDataFromScontrolOutput constructs a *TableData from raw scontrol
// one-liner output without executing an external command. It replicates the
// column-building logic from getScontrolDataWithTimeout so that model-level
// benchmarks can work offline using testdata files.
func buildTableDataFromScontrolOutput(output string, columns *[]config.ColumnConfig) *TableData {
	rawRows := parseScontrolOutput(output)
	rows := make([][]CellValue, 0, len(rawRows))
	for _, rawRow := range rawRows {
		row := make([]CellValue, len(*columns))
		for j := range *columns {
			col := &(*columns)[j]
			switch {
			case col.ComputedColumn && col.ComputationType == "ratio":
				components := strings.Split(col.RawName, "//")
				if len(components) == 2 {
					num := strings.TrimSpace(components[0])
					den := strings.TrimSpace(components[1])
					row[j] = NewRatioValue(
						parseTypedValue(safeGetFromMap(rawRow, num), GetFieldType(num)),
						parseTypedValue(safeGetFromMap(rawRow, den), GetFieldType(den)),
					)
				} else {
					row[j] = NewStringValue("ERROR")
				}
			case col.DividedByColumn:
				parts := strings.Split(col.RawName, "//")
				vals := make([]string, len(parts))
				for i, p := range parts {
					vals[i] = safeGetFromMap(rawRow, p)
				}
				row[j] = parseTypedValue(strings.Join(vals, " / "), TypeString)
			default:
				row[j] = parseTypedValue(safeGetFromMap(rawRow, col.DisplayName), GetFieldType(col.DisplayName))
			}
		}
		rows = append(rows, row)
	}
	return &TableData{
		Headers:             columns,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows),
	}
}

// BenchmarkParseScontrolOutput_Nodes measures parsing the full 8 888-node nodes.txt.
func BenchmarkParseScontrolOutput_Nodes(b *testing.B) {
	raw, err := os.ReadFile(filepath.Join("testdata", "nodes.txt"))
	if err != nil {
		b.Fatalf("read nodes.txt: %v", err)
	}
	data := string(raw)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseScontrolOutput(data)
	}
}

// BenchmarkParseScontrolOutput_Jobs measures parsing jobs.txt (630 jobs).
func BenchmarkParseScontrolOutput_Jobs(b *testing.B) {
	raw, err := os.ReadFile(filepath.Join("testdata", "jobs.txt"))
	if err != nil {
		b.Fatalf("read jobs.txt: %v", err)
	}
	data := string(raw)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseScontrolOutput(data)
	}
}

// BenchmarkApplyFilters_Nodes measures ApplyFilters with an all-pass filter
// over the 8 888-node dataset (worst case: no rows dropped, all iterated).
func BenchmarkApplyFilters_Nodes(b *testing.B) {
	raw, err := os.ReadFile(filepath.Join("testdata", "nodes.txt"))
	if err != nil {
		b.Fatalf("read nodes.txt: %v", err)
	}
	td := buildTableDataFromScontrolOutput(string(raw), config.NodeViewColumns)
	filters := map[int]string{
		config.NodeViewColumnsStateIndex:     config.ALL_CATEGORIES_OPTION,
		config.NodeViewColumnsPartitionIndex: config.ALL_CATEGORIES_OPTION,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = td.ApplyFilters(filters)
	}
}

// BenchmarkConvertRowsToSingleStrings measures the string-join step over 8 888 rows.
func BenchmarkConvertRowsToSingleStrings(b *testing.B) {
	raw, err := os.ReadFile(filepath.Join("testdata", "nodes.txt"))
	if err != nil {
		b.Fatalf("read nodes.txt: %v", err)
	}
	td := buildTableDataFromScontrolOutput(string(raw), config.NodeViewColumns)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertRowsToRowsAsSingleStrings(td.Rows)
	}
}

// BenchmarkDeepCopy_Nodes measures a full deep copy of a 8 888-row TableData.
func BenchmarkDeepCopy_Nodes(b *testing.B) {
	raw, err := os.ReadFile(filepath.Join("testdata", "nodes.txt"))
	if err != nil {
		b.Fatalf("read nodes.txt: %v", err)
	}
	td := buildTableDataFromScontrolOutput(string(raw), config.NodeViewColumns)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = td.DeepCopy()
	}
}

// BenchmarkApplyRegexSearchFilter measures regex compilation + row filtering over
// the full 8 888-node RowsAsSingleStrings slice, matching the hot-path in
// StuiView.ApplyRegexSearchFilterToRows.
func BenchmarkApplyRegexSearchFilter(b *testing.B) {
	raw, err := os.ReadFile(filepath.Join("testdata", "nodes.txt"))
	if err != nil {
		b.Fatalf("read nodes.txt: %v", err)
	}
	td := buildTableDataFromScontrolOutput(string(raw), config.NodeViewColumns)
	pattern := "(?i)IDLE"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re, _ := regexp.Compile(pattern)
		filteredRows := make([][]CellValue, 0, len(td.Rows)/2)
		for j, s := range td.RowsAsSingleStrings {
			if re.MatchString(s) {
				filteredRows = append(filteredRows, td.Rows[j])
			}
		}
		_ = filteredRows
	}
}
