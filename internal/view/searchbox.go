package view

import (
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupSearchBox() {
	a.SearchBox = tview.NewInputField().
		SetLabel("  Regex search (case-insensitive): ").
		SetLabelColor(tcell.ColorDarkOrange).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			a.SearchPattern = strings.TrimSpace(text)
			a.SearchActive = a.SearchPattern != ""

			// Cancel any pending updates
			if a.searchTimer != nil {
				a.searchTimer.Stop()
			}

			// Schedule new update after delay
			a.searchTimer = time.AfterFunc(config.SearchDebounceInterval, func() {
				a.App.QueueUpdateDraw(func() {
					if a.CurrentTableView != nil {
						a.RerenderTableView(a.CurrentTableView)
					}
				})
			})
		})
	a.SearchBox.SetBorder(false)

	// Set up input capture for search box
	a.SearchBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.HideSearchBox()
			a.RerenderTableView(a.CurrentTableView)
			return nil
		case tcell.KeyEnter:
			if a.SearchPattern == "" {
				a.HideSearchBox()
			} else {
				a.App.SetFocus(a.CurrentTableView)
			}
			return nil
		}
		return event
	})
}

func (a *App) ShowSearchBox(grid *tview.Grid) {
	if a.CurrentTableView == nil {
		return
	}

	// Clear and rebuild the grid with search box
	grid.Clear()
	grid.SetRows(1, 0)                                       // 1 row for search, rest for table
	grid.AddItem(a.SearchBox, 0, 0, 1, 1, 0, 0, false)       // Don't focus by default
	grid.AddItem(a.CurrentTableView, 1, 0, 1, 1, 0, 0, true) // Keep table focused

	a.SearchActive = true
}

func (a *App) HideSearchBox() {
	// Stop any pending search updates
	if a.searchTimer != nil {
		a.searchTimer.Stop()
		a.searchTimer = nil
	}

	if a.CurrentTableView == nil {
		return
	}

	// Get the appropriate grid
	// Gross, search bar should not need to figure this out..
	grid := a.NodeGrid
	if a.CurrentTableView == a.JobsView {
		grid = a.JobGrid
	} else if a.CurrentTableView == a.AcctView {
		grid = a.AcctGrid
	}

	// Clear and rebuild grid without search box
	grid.Clear()
	grid.SetRows(0) // Just table
	grid.AddItem(a.CurrentTableView, 0, 0, 1, 1, 0, 0, true)

	// Reset search state
	a.SearchBox.SetText("")
	a.SearchActive = false
	a.App.SetFocus(a.CurrentTableView)
}
