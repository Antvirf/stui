package view

import (
	"log"
	"slices"
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

		// Don't allow pane switching when prompts are open or selectors are in focus
		if a.CommandModalOpen ||
			a.SearchBox.HasFocus() ||
			a.PartitionSelector.HasFocus() ||
			a.SacctMgrEntitySelector.HasFocus() {
			return event
		}

		switch event.Rune() {
		case '?':
			a.ShowModalPopup("Shortcuts", config.KEYBOARD_SHORTCUTS)
		case '1':
			a.SwitchToPage(NODES_PAGE)
			a.CurrentTableView = a.NodesView.Table
			a.SetHeaderGridInnerContents(
				a.PartitionSelector,
				a.NodeStateSelector,
			)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.NodesView.Grid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.NodesView.Table)
			a.NodesView.Render()
			go a.NodesView.FetchAndRenderIfStale(config.RefreshInterval)
			return nil
		case '2':
			a.SwitchToPage(JOBS_PAGE)
			a.CurrentTableView = a.JobsView.Table
			a.SetHeaderGridInnerContents(
				a.PartitionSelector,
				a.JobStateSelector,
			)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.JobsView.Grid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.JobsView.Table)
			a.JobsView.Render()
			go a.JobsView.FetchAndRenderIfStale(config.RefreshInterval)
			return nil
		case '3':
			a.SwitchToPage(SDIAG_PAGE)
			a.PagesContainer.SetTitle(" Scheduler status (sdiag) ")
			a.CurrentTableView = nil
			a.HideSearchBox()
			a.SetHeaderGridInnerContents(tview.NewBox())
			a.UpdateHeaderLineOne("")
			a.UpdateHeaderLineTwo("")
			return nil
		case '4':
			if config.SacctEnabled {
				a.SwitchToPage(SACCTMGR_PAGE)
				a.CurrentTableView = a.SacctMgrView.Table
				a.SetHeaderGridInnerContents(a.SacctMgrEntitySelector)
				if a.SearchPattern != "" {
					a.ShowSearchBox(a.SacctMgrView.Grid)
				} else {
					a.HideSearchBox()
				}
				a.App.SetFocus(a.SacctMgrView.Table)
				a.SacctMgrView.Render()
				go a.SacctMgrView.FetchAndRenderIfStale(config.RefreshInterval)
			}
			return nil
		}
		return event
	})

	if config.SacctEnabled {
		a.SacctMgrView.Table.SetInputCapture(
			tableviewInputCapture(
				a,
				a.SacctMgrView.Table,
				&a.SacctMgrView.Selection,
				"",              // Used for command modal, ignored if blank
				func(string) {}, // Null func for detail view
			),
		)
	}

	// Table view keybinds
	a.NodesView.Table.SetInputCapture(
		tableviewInputCapture(
			a,
			a.NodesView.Table,
			&a.NodesView.Selection,
			"NodeName", // Used for command modal
			a.ShowNodeDetails,
		),
	)
	a.JobsView.Table.SetInputCapture(
		tableviewInputCapture(
			a,
			a.JobsView.Table,
			&a.JobsView.Selection,
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
		case a.NodesView.Table:
			data = a.NodesProvider.Data()
			grid = a.NodesView.Grid
		case a.JobsView.Table:
			data = a.JobsProvider.Data()
			grid = a.JobsView.Grid
		case a.SacctMgrView.Table:
			data = a.SacctMgrProvider.Data()
			grid = a.SacctMgrView.Grid
		}
		switch event.Rune() {
		case '/':
			a.ShowSearchBox(grid)
			a.RenderCurrentView()
			a.App.SetFocus(a.SearchBox) // Only focus search when / is pressed
			return nil
		case ' ':
			row, _ := view.GetSelection()
			// Certain tables in Sacctmgr have no clear ID, and the current selection implementation relies
			// on the first column of a row to be an identifier column.
			if view == a.SacctMgrView.Table &&
				slices.Contains(
					model.SACCTMGR_ENTITY_TABLES_WITH_NO_CLEAR_ID,
					config.SacctMgrCurrentEntity,
				) {
				return nil
			}
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
			if a.GetCurrentPageName() != SACCTMGR_PAGE {
				a.App.SetFocus(a.PartitionSelector)
			}
		case 'e':
			if a.GetCurrentPageName() == SACCTMGR_PAGE {
				a.App.SetFocus(a.SacctMgrEntitySelector)
			}
		case 's':
			switch a.GetCurrentPageName() {
			case NODES_PAGE:
				a.App.SetFocus(a.NodeStateSelector)
			case JOBS_PAGE:
				a.App.SetFocus(a.JobStateSelector)
			}
		case 'c':
			// This section is only active if there is a commandModalFilter specified.
			if commandModalFilter != "" {
				// If user has a selection, use the selection
				if len(*selection) > 0 {
					a.ShowCommandModal(commandModalFilter, *selection, a.GetCurrentPageName())
				} else {
					// Otherwise, try to use the current node under the cursor, if any
					row, _ := view.GetSelection()
					if row > 0 {
						a.ShowCommandModal(commandModalFilter, map[string]bool{
							view.GetCell(row, 0).Text: true,
						},
							a.GetCurrentPageName(),
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
						if row[0] == entryName {
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
				a.RenderCurrentView()
				return nil
			}
		}
		return event
	}
}
