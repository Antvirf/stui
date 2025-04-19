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
		SetLabelColor(searchboxLabelColor).
		SetFieldBackgroundColor(dropdownBackgroundColor).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			a.SearchPattern = strings.TrimSpace(text)
			a.NodesView.SetSearchPattern(a.SearchPattern)
			a.JobsView.SetSearchPattern(a.SearchPattern)
			a.SacctMgrView.SetSearchPattern(a.SearchPattern)

			// Cancel any pending updates
			if a.searchTimer != nil {
				a.searchTimer.Stop()
			}

			// Schedule new update after delay
			a.searchTimer = time.AfterFunc(config.SearchDebounceInterval, func() {
				a.App.QueueUpdateDraw(func() {
					a.RenderCurrentView()
				})
			})
		})
	a.SearchBox.SetBorder(false)

	// Set up input capture for search box
	a.SearchBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.HideSearchBox()
			a.RenderCurrentView()
			return nil
		case tcell.KeyEnter:
			if a.SearchPattern == "" {
				a.HideSearchBox()
			} else {
				a.RenderCurrentView()
				a.App.SetFocus(*a.GetCurrentPage())
			}
			return nil
		}
		return event
	})
}

func (a *App) ShowSearchBox(grid *tview.Grid) {
	pageName, _ := a.Pages.GetFrontPage()
	var table *tview.Table
	switch pageName {
	case SDIAG_PAGE:
		return // No search on sdiag
	case NODES_PAGE:
		a.NodesView.SetSearchEnabled(true)
		table = a.NodesView.Table
	case JOBS_PAGE:
		a.JobsView.SetSearchEnabled(true)
		table = a.JobsView.Table
	case SACCTMGR_PAGE:
		a.SacctMgrView.SetSearchEnabled(true)
		table = a.SacctMgrView.Table
	}

	// Clear and rebuild the grid with search box
	// grid.Clear()
	grid.SetRows(1, 0)                                 // 1 row for search, rest for table
	grid.AddItem(a.SearchBox, 0, 0, 1, 1, 0, 0, false) // Don't focus by default
	grid.AddItem(table, 1, 0, 1, 1, 0, 0, true)        // Keep table focused
	a.SearchActive = true
}

func (a *App) HideSearchBox() {
	pageName, page := a.Pages.GetFrontPage()
	// TODO: This is really gross
	var grid *tview.Grid
	var table *tview.Table
	switch pageName {
	case SDIAG_PAGE:
		return // No search on sdiag
	case NODES_PAGE:
		a.NodesView.SetSearchEnabled(false)
		grid = a.NodesView.Grid
		table = a.NodesView.Table
	case JOBS_PAGE:
		a.JobsView.SetSearchEnabled(false)
		grid = a.JobsView.Grid
		table = a.JobsView.Table
	case SACCTMGR_PAGE:
		a.SacctMgrView.SetSearchEnabled(false)
		grid = a.SacctMgrView.Grid
		table = a.SacctMgrView.Table
	}

	// Stop any pending search updates
	if a.searchTimer != nil {
		a.searchTimer.Stop()
		a.searchTimer = nil
	}

	// Clear and rebuild grid without search box
	grid.Clear()
	grid.SetRows(0) // Just table
	grid.AddItem(table, 0, 0, 1, 1, 0, 0, true)

	// Reset search state
	a.SearchBox.SetText("")
	a.SearchActive = false
	a.App.SetFocus(page)
}
