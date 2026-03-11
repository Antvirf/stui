package view

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// GetVisibleRows returns the currently visible rows after all filters, search, and sorting
// have been applied - exactly as the user sees them in the table.
func (s *StuiView) GetVisibleRows() ([]string, [][]string) {
	if s.data == nil {
		return nil, nil
	}

	// Extract headers
	headers := make([]string, len(*s.data.Headers))
	for i, h := range *s.data.Headers {
		headers[i] = h.DisplayName
	}

	// Apply search filter (same logic as Render)
	filteredRows := s.data.Rows
	if s.searchEnabled && *s.searchPattern != "" {
		searchPatternRegex, err := regexp.Compile("(?i)" + *s.searchPattern)
		if err == nil {
			filteredRows = make([][]model.CellValue, 0, len(s.data.Rows)/2)
			for i, row := range s.data.Rows {
				if searchPatternRegex.MatchString(s.data.RowsAsSingleStrings[i]) {
					filteredRows = append(filteredRows, row)
				}
			}
		}
	}

	// Apply sort (same logic as Render)
	if s.sortColumn >= 0 && len(filteredRows) > 0 {
		sort.Slice(filteredRows, func(i, j int) bool {
			comparison := filteredRows[i][s.sortColumn].Compare(filteredRows[j][s.sortColumn])
			if s.sortDirection > 0 {
				return comparison < 0
			}
			return comparison > 0
		})
	}

	// Convert to string rows
	rows := make([][]string, len(filteredRows))
	for i, row := range filteredRows {
		rowStrings := make([]string, len(row))
		for j, cell := range row {
			rowStrings[j] = cell.Display()
		}
		rows[i] = rowStrings
	}

	return headers, rows
}

// ExportToCSV writes the given headers and rows to a CSV file at the specified path.
func ExportToCSV(filepath string, headers []string, rows [][]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// GenerateExportFilename returns a default filename for CSV export
func GenerateExportFilename() string {
	return fmt.Sprintf("stui-export-%s.csv", time.Now().Format("2006-01-02-150405"))
}

// ShowExportPrompt displays a modal with an input field for the export filename
func (a *App) ShowExportPrompt() {
	view := a.GetCurrentStuiView()
	if view == nil {
		a.ShowNotification("[red]Export not available for this view[white]", 2*time.Second)
		return
	}

	pageName := fmt.Sprintf("export-prompt-%d", time.Now().Unix())

	input := tview.NewInputField().
		SetLabel("Export filename: ").
		SetFieldStyle(
			tcell.StyleDefault.Background(rowCursorColorBackground),
		).
		SetText(GenerateExportFilename()).
		SetFieldWidth(0)

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(" Export current view as CSV (ESC to cancel) ").
			SetTextColor(generalTextColor),
			2, 0, false).
		AddItem(input, 1, 0, true)

	modal.SetBorder(true).
		SetBorderColor(modalBorderColor).
		SetBackgroundColor(generalBackgroundColor)

	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 4, false).
			AddItem(modal, 5, 0, true).
			AddItem(nil, 0, 4, false),
			0, 4, false).
		AddItem(nil, 0, 1, false)

	previousFocus := a.App.GetFocus()

	a.Pages.AddPage(pageName, centered, true, true)
	a.App.SetFocus(input)

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			filename := input.GetText()
			if filename == "" {
				return nil
			}

			headers, rows := view.GetVisibleRows()
			if headers == nil {
				a.Pages.RemovePage(pageName)
				a.App.SetFocus(previousFocus)
				a.ShowNotification("[red]No data to export[white]", 2*time.Second)
				return nil
			}

			err := ExportToCSV(filename, headers, rows)
			a.Pages.RemovePage(pageName)
			a.App.SetFocus(previousFocus)

			if err != nil {
				a.ShowNotification(
					fmt.Sprintf("[red]Export failed: %s[white]", err.Error()),
					3*time.Second,
				)
			} else {
				a.ShowNotification(
					fmt.Sprintf("[green]Exported %d rows to %s[white]", len(rows), filename),
					3*time.Second,
				)
			}
			return nil

		case tcell.KeyEsc:
			a.Pages.RemovePage(pageName)
			a.App.SetFocus(previousFocus)
			return nil
		}
		return event
	})
}
