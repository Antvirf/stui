package view

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.design/x/clipboard"
)

func (a *App) SetupKeybinds() {
	// Global keybinds (work anywhere except when typing in search)
	a.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if event.Key() == tcell.KeyCtrlC {
			a.App.Stop()
			duration := time.Since(a.startTime)
			rpm := float64(model.FetchCounter.Count) / duration.Seconds() * 60
			log.Printf(
				"Session stats: duration=%s, total_scheduler_calls=%d, requests_per_minute=%.1f",
				duration.Round(time.Second),
				model.FetchCounter.Count,
				rpm,
			)
			log.Print("Thank you for using stui!")
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
			a.FooterPaneLocation.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
			a.CurrentTableView = a.NodesView
			if a.SearchPattern != "" {
				a.ShowSearchBox()
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.NodesView)
			a.RerenderTableView(a.NodesView)
			return nil
		case '2':
			a.Pages.SwitchToPage("jobs")
			a.FooterPaneLocation.SetText("Nodes (1) - [::b]Jobs (2)[::-] - Scheduler (3)")
			a.CurrentTableView = a.JobsView
			if a.SearchPattern != "" {
				a.ShowSearchBox()
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.JobsView)
			a.RerenderTableView(a.JobsView)
			return nil
		case '3':
			a.Pages.SwitchToPage("scheduler")
			a.PagesContainer.SetTitle(" Scheduler status (sdiag) ")
			a.FooterPaneLocation.SetText("Nodes (1) - Jobs (2) - [::b]Scheduler (3)[::-]")
			a.CurrentTableView = nil
			a.HideSearchBox()
			return nil
		}
		return event
	})

	// Table view keybinds
	a.NodesView.SetInputCapture(
		tableviewInputCapture(
			a,
			a.NodesView,
			a.NodesTableData,
			&a.SelectedNodes,
			"NodeName", // Used for command modal
			a.ShowNodeDetails,
		),
	)
	a.JobsView.SetInputCapture(
		tableviewInputCapture(
			a,
			a.JobsView,
			a.JobsTableData,
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
	data *model.TableData,
	selection *map[string]bool,
	commandModalFilter string,
	detailsFunction func(string),
) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '/':
			a.ShowSearchBox()
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
						view.GetCell(row, col).SetBackgroundColor(tcell.ColorBlack)
					}
				} else {
					(*selection)[entryName] = true
					// Set all cells in row to orange background
					for col := 0; col < view.GetColumnCount(); col++ {
						view.GetCell(row, col).SetBackgroundColor(selectionColor)
					}
				}
			}
			return nil
		case 'p':
			a.App.SetFocus(a.PartitionSelector)
		case 'c':
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
			return nil
		case 'y':
			if len(*selection) > 0 {
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
				clipboard.Write(clipboard.FmtText, []byte(sb.String()))
				count := len(*selection)
				a.ShowNotification(
					fmt.Sprintf("[green]Copied %d selected row(s) to clipboard[white]", count),
					2*time.Second,
				)
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
