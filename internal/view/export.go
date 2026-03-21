package view

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ExportToCSV writes headers and rows to a CSV file at the specified path.
func ExportToCSV(filepath string, headers []string, rows [][]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}
	}
	return nil
}

// ShowExportPrompt displays a prompt asking for the export filename.
// Export is only supported on table-based panes (nodes, jobs, sacct, sacctmgr).
func (a *App) ShowExportPrompt() {
	currentPage := a.GetCurrentPageName()
	if !slices.Contains([]string{config.NODES_PAGE, config.JOBS_PAGE, config.SACCT_PAGE, config.SACCTMGR_PAGE}, currentPage) {
		return
	}

	view := a.GetCurrentStuiView()
	if view == nil {
		return
	}

	pageName := fmt.Sprintf("export-prompt-%d", time.Now().Unix())
	defaultFilename := fmt.Sprintf("stui-export-%s-%s.csv", a.GetCurrentPageName(), time.Now().Format("2006-01-02-150405"))

	input := tview.NewInputField().
		SetLabel("Export filename: ").
		SetFieldStyle(tcell.StyleDefault.Background(rowCursorColorBackground)).
		SetText(defaultFilename).
		SetFieldWidth(0)

	modal := tview.NewFlex().SetDirection(tview.FlexRow).
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

	closePrompt := func() {
		a.Pages.RemovePage(pageName)
		a.App.SetFocus(previousFocus)
	}

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			filename := input.GetText()
			if filename == "" {
				return nil
			}

			headers := view.GetHeadersAsText()
			rows := view.GetVisibleRowsAsText()

			closePrompt()
			if headers == nil {
				a.ShowNotification("[red]No data to export[white]", 2*time.Second)
				return nil
			}
			if err := ExportToCSV(filename, headers, rows); err != nil {
				a.ShowNotification(fmt.Sprintf("[red]Export failed: %s[white]", err), 3*time.Second)
			} else {
				a.ShowNotification(fmt.Sprintf("[green]Exported %d rows to %s[white]", len(rows), filename), 3*time.Second)
			}
			return nil
		case tcell.KeyEsc:
			closePrompt()
			return nil
		}
		return event
	})
}
