package view

import (
	"log"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupKeybinds() {
	// Global keybinds (work anywhere except when typing in search)
	a.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if event.Key() == tcell.KeyCtrlC {
			a.App.Stop()
			if !config.Quiet {
				duration := time.Since(a.startTime)
				rpm := float64(model.FetchCounter.Count) / duration.Seconds() * 60
				rpm = min(rpm, float64(model.FetchCounter.Count))
				log.Printf(
					"END: Session stats: duration=%s, total_scheduler_calls=%d, requests_per_minute=%.1f",
					duration.Round(time.Second),
					model.FetchCounter.Count,
					rpm,
				)
				log.Print("Thank you for using stui!")
			}
			return event
		}

		// Don't allow pane switching while typing in search
		// same if command prompt is open
		if a.SearchBox.HasFocus() || a.CommandModalOpen {
			return event
		}

		switch event.Rune() {
		case '?':
			a.ShowModalPopup("Shortcuts", config.KEYBOARD_SHORTCUTS)
		case '1':
			a.Pages.SwitchToPage("nodes")
			a.setActiveTab("nodes")
			a.CurrentTableView = a.NodesView
			a.SetHeaderGridInnerContents(a.PartitionSelector)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.NodeGrid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.NodesView)
			a.RerenderTableView(a.NodesView)
			return nil
		case '2':
			a.Pages.SwitchToPage("jobs")
			a.setActiveTab("jobs")
			a.CurrentTableView = a.JobsView
			a.SetHeaderGridInnerContents(a.PartitionSelector)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.JobGrid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.JobsView)
			a.RerenderTableView(a.JobsView)
			return nil
		case '3':
			a.Pages.SwitchToPage("scheduler")
			a.PagesContainer.SetTitle(" Scheduler status (sdiag) ")
			a.setActiveTab("scheduler")
			a.CurrentTableView = nil
			a.HideSearchBox()
			a.SetHeaderGridInnerContents(tview.NewBox())
			return nil
		case '4':
			if config.SacctEnabled {

				a.Pages.SwitchToPage("accounting")
				a.setActiveTab("accounting")
				a.CurrentTableView = a.SacctMgrView
				a.SetHeaderGridInnerContents(a.SacctMgrEntitySelector)
				if a.SearchPattern != "" {
					a.ShowSearchBox(a.AcctGrid)
				} else {
					a.HideSearchBox()
				}
				a.App.SetFocus(a.SacctMgrView)
				a.RerenderTableView(a.SacctMgrView)
			}
			return nil
		}
		return event
	})

	if config.SacctEnabled {
		a.SacctMgrView.SetInputCapture(
			tableviewInputCapture(
				a,
				a.SacctMgrView,
				&a.SelectedAcctRows,
				"",              // Used for command modal, ignored if blank
				func(string) {}, // Null func for detail view
			),
		)
	}

	// Table view keybinds
	a.NodesView.SetInputCapture(
		tableviewInputCapture(
			a,
			a.NodesView,
			&a.SelectedNodes,
			"NodeName", // Used for command modal
			a.ShowNodeDetails,
		),
	)
	a.JobsView.SetInputCapture(
		tableviewInputCapture(
			a,
			a.JobsView,
			&a.SelectedJobs,
			"JobId", // Used for command modal
			a.ShowJobDetails,
		),
	)
}

// Handles all inputs for table views (nodes and jobs)
func tableviewInputCapture(
	a *App,
	view *tview.Table,
	selection *map[string]bool,
	commandModalFilter string,
	detailsFunction func(string),
) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		// Get current table data based on which view we're in
		// Passing this as a pointer will cause a nil pointer dereference
		var data *model.TableData
		var grid *tview.Grid
		switch view {
		case a.NodesView:
			data = a.NodesTableData
			grid = a.NodeGrid
		case a.JobsView:
			data = a.JobsTableData
			grid = a.JobGrid
		case a.SacctMgrView:
			data = a.AcctTableData
			grid = a.AcctGrid
		}
		switch event.Rune() {
		case '/':
			a.ShowSearchBox(grid)
			a.RerenderTableView(view)
			a.App.SetFocus(a.SearchBox) // Only focus search when / is pressed
			return nil
		case ' ':
			row, _ := view.GetSelection()
			if row > 0 { // Skip header row
				entryName := view.GetCell(row, 0).Text
				if (*selection)[entryName] {
					delete(*selection, entryName)
					// Set all cells in row to default background
					for col := 0; col < view.GetColumnCount(); col++ {
						view.GetCell(row, col).
							SetBackgroundColor(generalBackgroundColor).
							SetSelectedStyle(
								tcell.StyleDefault.
									Background(rowCursorColorBackground),
							)
					}
				} else {
					(*selection)[entryName] = true
					// Set all cells in row to orange background
					for col := 0; col < view.GetColumnCount(); col++ {
						view.GetCell(row, col).
							SetBackgroundColor(selectionColor).
							SetSelectedStyle(
								tcell.StyleDefault.
									Background(selectionHighlightColor),
							)
					}
				}
			}
			return nil
		case 'p':
			// TODO: This is gross, fix
			if a.CurrentTableView != a.SacctMgrView {
				a.App.SetFocus(a.PartitionSelector)
			}
		case 'e':
			// TODO: This is gross, fix
			if a.CurrentTableView == a.SacctMgrView {
				a.App.SetFocus(a.SacctMgrEntitySelector)
			}
		case 'c':
			// This section is only active if there is a commandModalFilter specified.
			if commandModalFilter != "" {
				// If user has a selection, use the selection
				if len(*selection) > 0 {
					a.ShowCommandModal(commandModalFilter, *selection)
				} else {
					// Otherwise, try to use the current node under the cursor, if any
					row, _ := view.GetSelection()
					if row > 0 {
						a.ShowCommandModal(commandModalFilter, map[string]bool{
							view.GetCell(row, 0).Text: true,
						},
						)
					}
				}
			}
			return nil
		case 'y':
			if len(*selection) > 0 && data != nil {
				var sb strings.Builder
				for entryName := range *selection {
					// Find the node in our table data
					for _, row := range data.Rows {
						if row[0] == entryName { // NodeName is first column
							if config.CopyFirstColumnOnly {
								sb.WriteString(row[0])
							} else {
								sb.WriteString(strings.Join(row, " "))
							}
							sb.WriteString(config.CopiedLinesSeparator)
							break
						}
					}
				}
				a.copyToClipBoard(sb.String())
				return nil
			}
		}

		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := view.GetSelection()
			if row > 0 { // Skip header row
				entryName := view.GetCell(row, 0).Text
				detailsFunction(entryName)
				return nil
			}
		case tcell.KeyEsc:
			if a.SearchActive {
				a.HideSearchBox()
				a.RerenderTableView(view)
				return nil
			}
		}
		return event
	}
}
