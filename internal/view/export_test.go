package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportToCSV(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("basic export", func(t *testing.T) {
		filepath := filepath.Join(tmpDir, "test-basic.csv")
		headers := []string{"Name", "State", "CPU"}
		rows := [][]string{
			{"node01", "IDLE", "0.5"},
			{"node02", "ALLOCATED", "4.2"},
		}

		err := ExportToCSV(filepath, headers, rows)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		assert.Len(t, lines, 3)
		assert.Equal(t, "Name,State,CPU", lines[0])
		assert.Equal(t, "node01,IDLE,0.5", lines[1])
		assert.Equal(t, "node02,ALLOCATED,4.2", lines[2])
	})

	t.Run("empty rows", func(t *testing.T) {
		filepath := filepath.Join(tmpDir, "test-empty.csv")
		headers := []string{"Name", "State"}
		rows := [][]string{}

		err := ExportToCSV(filepath, headers, rows)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		assert.Len(t, lines, 1)
		assert.Equal(t, "Name,State", lines[0])
	})

	t.Run("csv special characters are escaped", func(t *testing.T) {
		filepath := filepath.Join(tmpDir, "test-special.csv")
		headers := []string{"Name", "Reason"}
		rows := [][]string{
			{"node01", "user requested, maintenance"},
			{"node02", `with "quotes"`},
		}

		err := ExportToCSV(filepath, headers, rows)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath)
		require.NoError(t, err)

		// CSV library handles quoting for commas and quotes
		assert.Contains(t, string(content), `"user requested, maintenance"`)
		assert.Contains(t, string(content), `"with ""quotes"""`)
	})

	t.Run("invalid path returns error", func(t *testing.T) {
		err := ExportToCSV("/nonexistent/dir/test.csv", []string{"A"}, [][]string{})
		assert.Error(t, err)
	})
}

func TestGetVisibleRowsAsText(t *testing.T) {
	// Create a minimal StuiView with test data
	searchPattern := ""
	headers := &[]config.ColumnConfig{
		{RawName: "Name", DisplayName: "Name"},
		{RawName: "State", DisplayName: "State"},
	}

	view := &StuiView{
		searchEnabled: false,
		searchPattern: &searchPattern,
		sortColumn:    -1,
		sortDirection: SORT_NONE,
		data: &model.TableData{
			Headers: headers,
			Rows: [][]model.CellValue{
				{model.NewStringValue("node01"), model.NewStringValue("IDLE")},
				{model.NewStringValue("node02"), model.NewStringValue("DOWN")},
				{model.NewStringValue("node03"), model.NewStringValue("IDLE")},
			},
			RowsAsSingleStrings: []string{"node01IDLE", "node02DOWN", "node03IDLE"},
		},
	}

	t.Run("no filters returns all rows", func(t *testing.T) {
		hdrs := view.GetHeadersAsText()
		rows := view.GetVisibleRowsAsText()
		assert.Equal(t, []string{"Name", "State"}, hdrs)
		assert.Len(t, rows, 3)
		assert.Equal(t, "node01", rows[0][0])
	})

	t.Run("with search filter", func(t *testing.T) {
		searchPattern = "DOWN"
		view.searchEnabled = true

		hdrs := view.GetHeadersAsText()
		rows := view.GetVisibleRowsAsText()
		assert.Equal(t, []string{"Name", "State"}, hdrs)
		assert.Len(t, rows, 1)
		assert.Equal(t, "node02", rows[0][0])
		assert.Equal(t, "DOWN", rows[0][1])

		// Reset
		searchPattern = ""
		view.searchEnabled = false
	})

	t.Run("nil data returns nil", func(t *testing.T) {
		nilView := &StuiView{data: nil}
		hdrs := nilView.GetHeadersAsText()
		rows := nilView.GetVisibleRowsAsText()
		assert.Nil(t, hdrs)
		assert.Nil(t, rows)
	})

	t.Run("with sort ascending", func(t *testing.T) {
		view.sortColumn = 0
		view.sortDirection = SORT_ASC

		rows := view.GetVisibleRowsAsText()
		assert.Len(t, rows, 3)
		assert.Equal(t, "node01", rows[0][0])
		assert.Equal(t, "node02", rows[1][0])
		assert.Equal(t, "node03", rows[2][0])

		view.sortColumn = -1
		view.sortDirection = SORT_NONE
	})

	t.Run("with sort descending", func(t *testing.T) {
		view.sortColumn = 0
		view.sortDirection = SORT_DESC

		rows := view.GetVisibleRowsAsText()
		assert.Len(t, rows, 3)
		assert.Equal(t, "node03", rows[0][0])
		assert.Equal(t, "node02", rows[1][0])
		assert.Equal(t, "node01", rows[2][0])

		view.sortColumn = -1
		view.sortDirection = SORT_NONE
	})
}
