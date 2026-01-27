package view

import (
	"fmt"
	"regexp"
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
			a.ShowModalPopupString(
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
			a.PagesContainer.SetTitle(a.NodesView.completeTitle)
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
			a.PagesContainer.SetTitle(a.JobsView.completeTitle)
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
				a.PagesContainer.SetTitle(a.SacctView.completeTitle)
				go a.App.QueueUpdateDraw(func() {
					a.SacctView.FetchIfStaleAndRender(config.RefreshInterval)
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
				a.PagesContainer.SetTitle(a.SacctMgrView.completeTitle)
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
				"", // Used for command modal, ignored if blank
				a.ShowSacctJobDetails,
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
			selectRow(a, view, row, selection, data.Length(), false)
			// Hacky way to update selection counter immediately
			a.PagesContainer.SetTitle(
				hackyUpdateTitleWithSelectionCount(a.PagesContainer.GetTitle(), selection),
			)
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
						entryName := strings.TrimSpace(view.GetCell(row, 0).Text)
						if entryName != "" {
							a.ShowStandardCommandModal(commandModalFilter, map[string]bool{
								entryName: true,
							},
								a.GetCurrentPageName(),
							)
						}
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
						if row[0].Display() == entryName {
							if config.CopyFirstColumnOnly {
								sb.WriteString(row[0].Display())
							} else {
								// Convert row to strings
								rowStrings := make([]string, len(row))
								for i, cell := range row {
									rowStrings[i] = cell.Display()
								}
								sb.WriteString(strings.Join(rowStrings, " "))
							}
							sb.WriteString(config.CopiedLinesSeparator)
							break
						}
					}
				}
				a.copyToClipBoard(sb.String(), "[green]Copied row details clipboard[white]")
				return nil
			}
		}

		switch event.Key() {
		case tcell.KeyCtrlA:
			rows := view.GetRowCount()
			for rowIndex := 1; rowIndex < rows; rowIndex++ {
				selectRow(a, view, rowIndex, selection, data.Length(), true)
			}
			a.ShowNotification("[green]Ctrl+A: Select all visible rows[white]", 2*time.Second)

			// Hacky way to update selection counter immediately
			a.PagesContainer.SetTitle(
				hackyUpdateTitleWithSelectionCount(a.PagesContainer.GetTitle(), selection),
			)

			return nil
		case tcell.KeyCtrlI:
			rows := view.GetRowCount()
			for rowIndex := 1; rowIndex < rows; rowIndex++ {
				selectRow(a, view, rowIndex, selection, data.Length(), false)
			}
			// Hacky way to update selection counter immediately
			a.PagesContainer.SetTitle(
				hackyUpdateTitleWithSelectionCount(a.PagesContainer.GetTitle(), selection),
			)
			a.ShowNotification("[green]Ctrl+I: Invert selection[white]", 2*time.Second)
		case tcell.KeyEnter:
			row, _ := view.GetSelection()
			if row > 0 { // Skip header row
				entryName := strings.TrimSpace(view.GetCell(row, 0).Text)
				if entryName != "" {
					detailsFunction(entryName)
				}
				return nil
			}
		case tcell.KeyCtrlR:
			// Manual refresh of currently visible view
			a.optionalRefreshAndRenderCurrentView(true)
			a.ShowNotification("[green]Ctrl+R: Manual data refresh[white]", 2*time.Second)
		case tcell.KeyCtrlD:
			row, _ := view.GetSelection()
			if row == 0 { // Skip if user is on header row / there is on data
				return nil
			}
			// The below is an ugly way to check that we're in the jobs view
			if strings.Contains(commandModalFilter, "JobId") {
				SCANCEL_COMMAND := "scancel "
				// If user has a selection, use the selection
				if len(*selection) > 0 {
					a.ShowStandardCommandModal(SCANCEL_COMMAND, *selection, a.GetCurrentPageName())
				} else {
					// Otherwise, try to use the current node under the cursor, if any
					if row > 0 {
						entryName := strings.TrimSpace(view.GetCell(row, 0).Text)
						if entryName != "" {
							a.ShowStandardCommandModal(SCANCEL_COMMAND, map[string]bool{
								entryName: true,
							},
								a.GetCurrentPageName(),
							)
						}
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
				rowId := strings.TrimSpace(view.GetCell(row, 0).Text)
				if rowId != "" {
					a.ExecutePluginForShortcut(event.Key(), a.GetCurrentPageName(), rowId)
				}
			}
		}
		return event
	}
}

// This is a mega ugly way to do it. Some longer term refactoring needed here
// to async share messages between components.
func hackyUpdateTitleWithSelectionCount(title string, selection *map[string]bool) string {
	selectionCount := getSelectionCount(selection)

	// If it already contains '| %d selected', remove it
	re := regexp.MustCompile(`\|\s.*`)
	title = re.ReplaceAllString(title, "")

	if selectionCount == 0 {
		return title
	}
	return fmt.Sprintf("%s| %s selected", title, FormatNumberWithCommas(selectionCount))
}

func getSelectionCount(selection *map[string]bool) int {
	return len(*selection)
}

func selectRow(a *App, view *tview.Table, rowIndex int, selection *map[string]bool, dataLength int, additionalSelectOnly bool) {
	// Certain tables in Sacctmgr have no clear ID, and the current selection implementation relies
	// on the first column of a row to be an identifier column.
	if view == a.SacctMgrView.Table &&
		slices.Contains(
			model.SACCTMGR_ENTITY_TABLES_WITH_NO_CLEAR_ID,
			config.SacctMgrCurrentEntity,
		) {
		return
	}

	if rowIndex <= 0 { // Skip header and -ve rows
		return
	}

	if dataLength <= 0 {
		return
	}

	// Cells can be padded for width/alignment reasons, hence we have to trim.
	entryName := strings.TrimSpace(view.GetCell(rowIndex, 0).Text)
	if entryName == "" {
		return
	}
	_, isSelected := (*selection)[entryName]

	if isSelected && !additionalSelectOnly {
		delete(*selection, entryName) // Clear selection

		// Check whether we should give this row a special color based on its state field
		stateText := view.GetCell(rowIndex, config.NodeViewColumnsStateIndex).Text
		colorizedColor, shouldColorizeRow := GetStateColorMapping(stateText)

		for col := 0; col < view.GetColumnCount(); col++ {
			cell := view.GetCell(rowIndex, col)
			cell.SetBackgroundColor(generalBackgroundColor).
				SetSelectedStyle(tcell.StyleDefault.Background(rowCursorColorBackground))
			if shouldColorizeRow {
				cell.SetTextColor(colorizedColor)
			} else {
				cell.SetTextColor(generalTextColor)
			}
		}
	} else {
		(*selection)[entryName] = false // This boolean is NEVER used. Only presence in this map matters.

		for col := 0; col < view.GetColumnCount(); col++ {
			view.GetCell(rowIndex, col).
				SetBackgroundColor(selectionColor).
				SetTextColor(selectionTextColor).
				SetSelectedStyle(tcell.StyleDefault.Background(selectionHighlightColor))
		}
	}
}
