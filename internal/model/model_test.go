package model

import (
	"testing"

	"github.com/antvirf/stui/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableData_WithTypedValues(t *testing.T) {
	headers := []config.ColumnConfig{
		{RawName: "NodeName", DisplayName: "NodeName"},
		{RawName: "CPUTot", DisplayName: "CPUTot"},
		{RawName: "AllocMem", DisplayName: "AllocMem"},
	}

	rows := [][]CellValue{
		{
			NewStringValue("node1"),
			NewIntegerValue("64"),
			NewMemoryValue("128G"),
		},
		{
			NewStringValue("node2"),
			NewIntegerValue("32"),
			NewMemoryValue("64G"),
		},
	}

	td := &TableData{
		Headers:             &headers,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows, nil),
	}

	assert.Equal(t, 2, td.Length())
	assert.Equal(t, "node1", td.Rows[0][0].Display())
	assert.Equal(t, "64", td.Rows[0][1].Display())
}

func TestTableData_DeepCopy(t *testing.T) {
	headers := []config.ColumnConfig{
		{RawName: "CPUTot", DisplayName: "CPUTot"},
	}

	rows := [][]CellValue{
		{NewIntegerValue("10")},
		{NewIntegerValue("20")},
	}

	td := &TableData{
		Headers:             &headers,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows, nil),
	}

	copy := td.DeepCopy()

	assert.Equal(t, td.Length(), copy.Length())
	assert.Equal(t, td.Rows[0][0].Display(), copy.Rows[0][0].Display())

	// Verify it's a copy, not the same slice
	assert.NotSame(t, &td.Rows, &copy.Rows)
}

func TestTableData_ApplyFilters(t *testing.T) {
	headers := []config.ColumnConfig{
		{RawName: "NodeName", DisplayName: "NodeName"},
		{RawName: "State", DisplayName: "State"},
	}

	rows := [][]CellValue{
		{NewStringValue("node1"), NewStringValue("IDLE")},
		{NewStringValue("node2"), NewStringValue("ALLOCATED")},
		{NewStringValue("node3"), NewStringValue("IDLE")},
	}

	td := &TableData{
		Headers:             &headers,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows, nil),
	}

	// Filter by State = IDLE (assuming index 1 is state column)
	filtered := td.ApplyFilters(map[int]string{1: "IDLE"})

	assert.Equal(t, 2, filtered.Length())
	assert.Equal(t, "node1", filtered.Rows[0][0].Display())
	assert.Equal(t, "node3", filtered.Rows[1][0].Display())
}

func TestTableData_RowToMap(t *testing.T) {
	headers := []config.ColumnConfig{
		{RawName: "NodeName", DisplayName: "NodeName"},
		{RawName: "CPUTot", DisplayName: "CPUTot"},
	}

	row := []CellValue{
		NewStringValue("node1"),
		NewIntegerValue("64"),
	}

	td := &TableData{
		Headers: &headers,
	}

	rowMap := td.rowToMap(row)

	assert.Equal(t, "node1", rowMap["NodeName"])
	assert.Equal(t, "64", rowMap["CPUTot"])
}

func TestTableData_GetRowAsMapById(t *testing.T) {
	headers := []config.ColumnConfig{
		{RawName: "NodeName", DisplayName: "NodeName"},
		{RawName: "CPUTot", DisplayName: "CPUTot"},
	}

	rows := [][]CellValue{
		{NewStringValue("node1"), NewIntegerValue("64")},
		{NewStringValue("node2"), NewIntegerValue("32")},
	}

	td := &TableData{
		Headers:             &headers,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows, nil),
	}

	rowMap, err := td.GetRowAsMapById("node2")
	require.NoError(t, err)
	assert.Equal(t, "node2", rowMap["NodeName"])
	assert.Equal(t, "32", rowMap["CPUTot"])

	_, err = td.GetRowAsMapById("node999")
	assert.Error(t, err)
}

func TestConvertRowsToRowsAsSingleStrings(t *testing.T) {
	rows := [][]CellValue{
		{NewStringValue("node1"), NewIntegerValue("64")},
		{NewStringValue("node2"), NewIntegerValue("32")},
	}

	result := convertRowsToRowsAsSingleStrings(rows, nil)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, "node164", result[0])
	assert.Equal(t, "node232", result[1])
}

func TestTableData_WithNullValues(t *testing.T) {
	headers := []config.ColumnConfig{
		{RawName: "NodeName", DisplayName: "NodeName"},
		{RawName: "CPUTot", DisplayName: "CPUTot"},
	}

	rows := [][]CellValue{
		{NewStringValue("node1"), NewIntegerValue("64")},
		{NewStringValue("node2"), NewIntegerValue("N/A")}, // Null value
	}

	td := &TableData{
		Headers:             &headers,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows, nil),
	}

	assert.False(t, td.Rows[0][1].IsNull())
	assert.True(t, td.Rows[1][1].IsNull())
	assert.Equal(t, "N/A", td.Rows[1][1].Display())
}
