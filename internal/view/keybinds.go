package view

import "github.com/gdamore/tcell/v2"

func (a *App) SetupKeybinds() {
	// Global keybinds (work anywhere except when typing in search)
	a.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Don't allow pane switching while typing in search
		if a.SearchBox.HasFocus() {
			return event
		}

		switch event.Rune() {
		case '1':
			a.Pages.SwitchToPage("nodes")
			a.Footer.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
			a.CurrentTableView = a.NodesView
			if a.SearchPattern != "" {
				a.ShowSearchBox()
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.NodesView)
			return nil
		case '2':
			a.Pages.SwitchToPage("jobs")
			a.Footer.SetText("Nodes (1) - [::b]Jobs (2)[::-] - Scheduler (3)")
			a.CurrentTableView = a.JobsView
			if a.SearchPattern != "" {
				a.ShowSearchBox()
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.JobsView)
			return nil
		case '3':
			a.Pages.SwitchToPage("scheduler")
			a.PagesContainer.SetTitle(" Scheduler status (sdiag) ")
			a.Footer.SetText("Nodes (1) - Jobs (2) - [::b]Scheduler (3)[::-]")
			a.CurrentTableView = nil
			a.HideSearchBox()
			return nil
		}
		return event
	})

	// Table view keybinds
	a.NodesView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '/':
			a.ShowSearchBox()
			a.RerenderTableView(a.NodesView)
			a.App.SetFocus(a.SearchBox) // Only focus search when / is pressed
			return nil
		case ' ':
			row, _ := a.NodesView.GetSelection()
			if row > 0 { // Skip header row
				nodeName := a.NodesView.GetCell(row, 0).Text
				if a.SelectedNodes[nodeName] {
					delete(a.SelectedNodes, nodeName)
					// Set all cells in row to default background
					for col := 0; col < a.NodesView.GetColumnCount(); col++ {
						a.NodesView.GetCell(row, col).SetBackgroundColor(tcell.ColorBlack)
					}
				} else {
					a.SelectedNodes[nodeName] = true
					// Set all cells in row to orange background
					for col := 0; col < a.NodesView.GetColumnCount(); col++ {
						a.NodesView.GetCell(row, col).SetBackgroundColor(selectionColor)
					}
				}
			}
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := a.NodesView.GetSelection()
			if row > 0 { // Skip header row
				nodeName := a.NodesView.GetCell(row, 0).Text
				a.ShowNodeDetails(nodeName)
				return nil
			}
		case tcell.KeyEsc:
			if a.SearchActive {
				a.HideSearchBox()
				a.RerenderTableView(a.NodesView)
				return nil
			}
		}
		return event
	})

	a.JobsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '/':
			a.ShowSearchBox()
			a.RerenderTableView(a.JobsView)
			a.App.SetFocus(a.SearchBox) // Only focus search when / is pressed
			return nil
		case ' ':
			row, _ := a.JobsView.GetSelection()
			if row > 0 { // Skip header row
				jobID := a.JobsView.GetCell(row, 0).Text
				if a.SelectedJobs[jobID] {
					delete(a.SelectedJobs, jobID)
					// Set all cells in row to default background
					for col := 0; col < a.JobsView.GetColumnCount(); col++ {
						a.JobsView.GetCell(row, col).SetBackgroundColor(tcell.ColorBlack)
					}
				} else {
					a.SelectedJobs[jobID] = true
					// Set all cells in row to orange background
					for col := 0; col < a.JobsView.GetColumnCount(); col++ {
						a.JobsView.GetCell(row, col).SetBackgroundColor(selectionColor)
					}
				}
			}
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := a.JobsView.GetSelection()
			if row > 0 { // Skip header row
				jobID := a.JobsView.GetCell(row, 0).Text
				a.ShowJobDetails(jobID)
				return nil
			}
		case tcell.KeyEsc:
			if a.SearchActive {
				a.HideSearchBox()
				a.RerenderTableView(a.JobsView)
				return nil
			}
		}
		return event
	})

	// Search box keybinds
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

	// Node detail view keybinds are set in ShowNodeDetails()
}
