package view

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func init() {
	config.ComputeConfigurations()
}

// benchProvider is a minimal DataProvider[*model.TableData] for benchmarks.
type benchProvider struct {
	data *model.TableData
}

func (p *benchProvider) Length() int                      { return p.data.Length() }
func (p *benchProvider) Fetch() error                     { return nil }
func (p *benchProvider) FetchIfStale(time.Duration) error { return nil }
func (p *benchProvider) Data() *model.TableData           { return p.data.DeepCopy() }
func (p *benchProvider) FilteredData() *model.TableData   { return p.data }
func (p *benchProvider) LastUpdated() time.Time           { return time.Time{} }
func (p *benchProvider) LastError() error                 { return nil }

// buildSyntheticTableData creates a *model.TableData with nRows synthetic node rows
// shaped to match config.NodeViewColumns (set up by init above).
func buildSyntheticTableData(nRows int) *model.TableData {
	cols := config.NodeViewColumns
	nCols := len(*cols)

	states := []string{"IDLE", "ALLOCATED", "DOWN", "DRAINED", "MIXED"}
	partitions := []string{"general", "physics", "chemistry", "mathematics", "biology"}

	rows := make([][]model.CellValue, nRows)
	rowStrings := make([]string, nRows)

	for i := range rows {
		row := make([]model.CellValue, nCols)
		for j, col := range *cols {
			switch col.DisplayName {
			case "NodeName":
				row[j] = model.NewStringValue(fmt.Sprintf("node%05d", i))
			case "Partitions":
				row[j] = model.NewStringValue(partitions[i%len(partitions)])
			case "State":
				row[j] = model.NewStringValue(states[i%len(states)])
			default:
				row[j] = model.NewStringValue("N/A")
			}
		}
		rows[i] = row

		var sb strings.Builder
		for _, c := range row {
			sb.WriteString(c.Display())
		}
		rowStrings[i] = sb.String()
	}

	return &model.TableData{
		Headers:             cols,
		Rows:                rows,
		RowsAsSingleStrings: rowStrings,
	}
}

// newBenchStuiView creates a minimal StuiView suitable for render benchmarks.
func newBenchStuiView(td *model.TableData, searchPattern *string) *StuiView {
	sv := &StuiView{
		Table:         tview.NewTable(),
		provider:      &benchProvider{data: td},
		data:          td,
		sortColumn:    -1,
		sortDirection: SORT_NONE,
		searchEnabled: false,
		searchPattern: searchPattern,
		updateTitleFunction: func(string) *tview.Box {
			return nil
		},
		errorNotificationFunction:     func(string) {},
		dataStateNotificationFunction: func(string) {},
	}
	sv.Table.
		SetBorders(false).
		SetFixed(1, 0).
		SetSelectable(true, false)
	return sv
}

// BenchmarkStuiViewRender_Nodes measures a full Render() pass over 8 888 node rows.
func BenchmarkStuiViewRender_Nodes(b *testing.B) {
	td := buildSyntheticTableData(8888)
	searchPattern := ""
	sv := newBenchStuiView(td, &searchPattern)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sv.Render()
	}
}

// BenchmarkStuiViewRender_WithSearch measures Render() with an active search pattern
// (adds per-row regex matching and per-cell highlight tracking).
func BenchmarkStuiViewRender_WithSearch(b *testing.B) {
	td := buildSyntheticTableData(8888)
	searchPattern := "IDLE"
	sv := newBenchStuiView(td, &searchPattern)
	sv.searchEnabled = true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sv.Render()
	}
}

// BenchmarkApplySortingToRows measures sorting 8 888 rows by string column.
// A fresh copy of the row slice is made each iteration so that sort order
// does not degrade across iterations.
func BenchmarkApplySortingToRows(b *testing.B) {
	td := buildSyntheticTableData(8888)
	searchPattern := ""
	sv := newBenchStuiView(td, &searchPattern)
	sv.sortColumn = 0 // NodeName — string comparison
	sv.sortDirection = SORT_ASC
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := make([][]model.CellValue, len(td.Rows))
		copy(rows, td.Rows)
		sv.ApplySortingToRows(&rows)
	}
}

// BenchmarkColorizeTableCell measures the colorization function called once per
// visible cell in every Render() pass.
func BenchmarkColorizeTableCell(b *testing.B) {
	cell := tview.NewTableCell("test-value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Exercise all code paths: plain, selected, search-matched, colorized
		colorizeTableCell(cell, false, false, false, tcell.ColorWhite)
		colorizeTableCell(cell, true, false, false, tcell.ColorWhite)
		colorizeTableCell(cell, false, true, false, tcell.ColorWhite)
		colorizeTableCell(cell, false, false, true, tcell.ColorRed)
	}
}
