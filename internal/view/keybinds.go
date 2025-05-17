package view

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupKeybinds() {
	// Global keybinds (work anywhere except when typing in search)
	a.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {
		case tcell.KeyCtrlC:
			a.App.Stop()
			duration := time.Since(a.startTime)
			rpm := float64(model.FetchCounter.Count) / duration.Seconds() * 60
			rpm = min(rpm, float64(model.FetchCounter.Count))
			logger.Printf(
				"END: Session stats: duration=%s, total_scheduler_calls=%d, requests_per_minute=%.1f",
				duration.Round(time.Second),
				model.FetchCounter.Count,
				rpm,
			)
			logger.Printf("Thank you for using stui!")
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
			a.ShowModalPopup(
				"Shortcuts",
				fmt.Sprintf(
					"%s\n%s",
					config.KEYBOARD_SHORTCUTS,
					GetKeyboardShortcutHelperForPage(a.GetCurrentPageName()),
				),
			)
		case '1':
			a.SwitchToPage(NODES_PAGE)
			a.CurrentTableView = a.NodesView.Table
			a.SetHeaderGridInnerContents(
				a.PartitionSelector,
				a.NodeStateSelector,
				a.SortSelector,
			)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.NodesView.Grid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.NodesView.Table)
			a.setupSortSelectorOptions(a.NodesProvider, a.NodesView.sortColumn)
			a.PagesContainer.SetTitle(a.NodesView.titleHeader)
			go a.App.QueueUpdateDraw(func() {
				a.NodesView.FetchIfStaleAndRender(config.RefreshInterval)
			})
			return nil
		case '2':
			a.SwitchToPage(JOBS_PAGE)
			a.CurrentTableView = a.JobsView.Table
			a.SetHeaderGridInnerContents(
				a.PartitionSelector,
				a.JobStateSelector,
				a.SortSelector,
			)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.JobsView.Grid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.JobsView.Table)
			a.setupSortSelectorOptions(a.JobsProvider, a.JobsView.sortColumn)
			a.PagesContainer.SetTitle(a.JobsView.titleHeader)
			go a.App.QueueUpdateDraw(func() {
				a.JobsView.FetchIfStaleAndRender(config.RefreshInterval)
			})
			return nil
		case '3':
			if config.SacctEnabled {
				a.SwitchToPage(SACCT_PAGE)
				a.CurrentTableView = a.SacctView.Table
				a.SetHeaderGridInnerContents(
					a.PartitionSelector,
					a.JobStateSelector,
					a.SortSelector,
				)
				if a.SearchPattern != "" {
					a.ShowSearchBox(a.SacctView.Grid)
				} else {
					a.HideSearchBox()
				}
				a.App.SetFocus(a.SacctView.Table)
				a.setupSortSelectorOptions(a.SacctProvider, a.SacctView.sortColumn)
				a.PagesContainer.SetTitle(a.SacctView.titleHeader)
				go a.App.QueueUpdateDraw(func() {
					a.SacctView.Render()
				})
			}
			return nil
		case '4':
			if config.SacctEnabled {
				a.SwitchToPage(SACCTMGR_PAGE)
				a.CurrentTableView = a.SacctMgrView.Table
				a.SetHeaderGridInnerContents(
					a.SacctMgrEntitySelector,
					a.SortSelector,
				)
				if a.SearchPattern != "" {
					a.ShowSearchBox(a.SacctMgrView.Grid)
				} else {
					a.HideSearchBox()
				}
				a.App.SetFocus(a.SacctMgrView.Table)
				a.setupSortSelectorOptions(a.SacctMgrProvider, a.SacctMgrView.sortColumn)
				a.PagesContainer.SetTitle(a.SacctMgrView.titleHeader)
				go a.App.QueueUpdateDraw(func() {
					a.SacctMgrView.FetchIfStaleAndRender(config.RefreshInterval)
				})
			}
			return nil
		case '5':
			a.SwitchToPage(SDIAG_PAGE)
			a.PagesContainer.SetTitle(" Scheduler status (sdiag) ")
			a.CurrentTableView = nil
			a.HideSearchBox()
			a.SetHeaderGridInnerContents(tview.NewBox())
			a.UpdateHeaderLineOne("")
			a.UpdateHeaderLineTwo("")
			return nil
		}
		return event
	})

	if config.SacctEnabled {
		a.SacctView.Table.SetInputCapture(
			tableViewInputCapture(
				a,
				a.SacctView.Table,
				&a.SacctView.Selection,
				"",              // Used for command modal
				func(string) {}, // Null func for detail view
			),
		)
		a.SacctMgrView.Table.SetInputCapture(
			tableViewInputCapture(
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
		tableViewInputCapture(
			a,
			a.NodesView.Table,
			&a.NodesView.Selection,
			"scontrol update NodeName=", // Used for command modal
			a.ShowNodeDetails,
		),
	)
	a.JobsView.Table.SetInputCapture(
		tableViewInputCapture(
			a,
			a.JobsView.Table,
			&a.JobsView.Selection,
			"scontrol update JobId=", // Used for command modal
			a.ShowJobDetails,
		),
	)
}

// Handles all inputs for table views (nodes and jobs)
func tableViewInputCapture(
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
		case a.SacctView.Table:
			data = a.SacctProvider.Data()
			grid = a.SacctView.Grid
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
			if a.GetCurrentPageName() == JOBS_PAGE || a.GetCurrentPageName() == NODES_PAGE || a.GetCurrentPageName() == SACCT_PAGE {
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
			case SACCT_PAGE:
				a.App.SetFocus(a.JobStateSelector)
			}
		case 'o':
			if a.GetCurrentPageName() == NODES_PAGE ||
				a.GetCurrentPageName() == JOBS_PAGE ||
				a.GetCurrentPageName() == SACCT_PAGE ||
				a.GetCurrentPageName() == SACCTMGR_PAGE {
				a.App.SetFocus(a.SortSelector)
			}
			return nil
		case 'c':
			// This section is only active if there is a commandModalFilter specified.
			if commandModalFilter != "" {
				// If user has a selection, use the selection
				if len(*selection) > 0 {
					a.ShowStandardCommandModal(commandModalFilter, *selection, a.GetCurrentPageName())
				} else {
					// Otherwise, try to use the current node under the cursor, if any
					row, _ := view.GetSelection()
					if row > 0 {
						a.ShowStandardCommandModal(commandModalFilter, map[string]bool{
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
		case tcell.KeyCtrlD:
			// The below is an ugly way to check that we're in the jobs view
			if strings.Contains(commandModalFilter, "JobId") {
				SCANCEL_COMMAND := "scancel "
				// If user has a selection, use the selection
				if len(*selection) > 0 {
					a.ShowStandardCommandModal(SCANCEL_COMMAND, *selection, a.GetCurrentPageName())
				} else {
					// Otherwise, try to use the current node under the cursor, if any
					row, _ := view.GetSelection()
					if row > 0 {
						a.ShowStandardCommandModal(SCANCEL_COMMAND, map[string]bool{
							view.GetCell(row, 0).Text: true,
						},
							a.GetCurrentPageName(),
						)
					}
				}
			}
			return nil
		case tcell.KeyEsc:
			if a.SearchActive {
				a.HideSearchBox()
				a.RenderCurrentView()
				return nil
			}
		default:
			// In case nothing else matched, perhaps its defined in a plugin.
			// Get the current row and pass it in.
			row, _ := view.GetSelection()
			if row > 0 {
				rowId := view.GetCell(row, 0).Text
				a.ExecutePluginForShortcut(event.Key(), a.GetCurrentPageName(), rowId)
			}
		}
		return event
	}
}
