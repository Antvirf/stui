package view

import (
	"fmt"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) GetCurrentPageName() string {
	page, _ := a.Pages.GetFrontPage()
	return page
}

func (a *App) GetCurrentPage() *tview.Primitive {
	_, page := a.Pages.GetFrontPage()
	return &page
}

func (a *App) SwitchToPage(pageName string) {
	a.setActiveTab(pageName)
	a.Pages.SwitchToPage(pageName)
}

func (a *App) RefreshAndRenderCurrentView() {
	a.optionalRefreshAndRenderCurrentView(true)
}

func (a *App) RefreshAndRenderPage(pageName string) {
	a.optionalRefreshAndRenderPage(pageName, true)
}

func (a *App) RenderCurrentView() {
	a.optionalRefreshAndRenderCurrentView(false)
}

func (a *App) optionalRefreshAndRenderCurrentView(refresh bool) {
	a.optionalRefreshAndRenderPage(
		a.GetCurrentPageName(),
		refresh,
	)
}

func (a *App) optionalRefreshAndRenderPage(pageName string, refresh bool) {
	switch pageName {
	case NODES_PAGE:
		if refresh {
			a.NodesProvider.Fetch()
		}
		a.NodesView.SetFilter(config.PartitionFilter)
		a.NodesView.Render()
	case JOBS_PAGE:
		if refresh {
			a.JobsProvider.Fetch()
		}
		a.JobsView.SetFilter(config.PartitionFilter)
		a.JobsView.Render()
	case SACCTMGR_PAGE:
		if refresh {
			a.SacctMgrProvider.Fetch()
		}
		a.SacctMgrView.Render()
	case SDIAG_PAGE:
		if refresh {
			d := a.SdiagProvider.Data()
			a.SchedView.SetText(d.Data)
		}
		// No rendering operation needed, TextView just gets its data set periodically
	}
	go a.App.QueueUpdateDraw(func() {})
}

func (a *App) ShowModalPopup(title, details string) {
	// Create new modal components each time (don't reuse)
	detailView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true). // Enable text wrapping
		SetTextAlign(tview.AlignLeft)
	detailView.SetText(details)

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(fmt.Sprintf(" %s (ESC to close) ", title)).
			SetTextColor(generalTextColor),
			2, 0, false).
		AddItem(detailView, 0, 1, true)

	modal.SetBorder(true).
		SetBorderColor(modalBorderColor).
		SetBackgroundColor(generalBackgroundColor)

	// Create centered container with fixed size (80% width, 90% height)
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(modal, 0, 10, true), // Increased height
			0, 16, false). // Increased width
		AddItem(nil, 0, 1, false)

	// Store current page before showing modal
	previousPageName, _ := a.Pages.GetFrontPage()
	previousFocus := a.App.GetFocus()

	// Add as overlay without switching pages
	pageName := "detailView"
	a.Pages.AddPage(pageName, centered, true, true)
	a.App.SetFocus(detailView)

	// Set up handler to return to correct view when closed
	detailView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.Pages.RemovePage(pageName)
			a.Pages.SwitchToPage(previousPageName)
			a.App.SetFocus(previousFocus)
			return nil
		}
		switch event.Rune() {
		case 'y':
			a.copyToClipBoard(details)
			return nil
		}
		return event
	})
}

func (a *App) setActiveTab(active string) {
	// Reset all to black
	a.TabNodesBox.SetBackgroundColor(generalBackgroundColor)
	a.TabJobsBox.SetBackgroundColor(generalBackgroundColor)
	a.TabSchedulerBox.SetBackgroundColor(generalBackgroundColor)
	a.TabAccountingBox.SetBackgroundColor(generalBackgroundColor)

	// Set active color
	switch active {
	case NODES_PAGE:
		a.TabNodesBox.SetBackgroundColor(paneSelectorHighlightColor)
	case JOBS_PAGE:
		a.TabJobsBox.SetBackgroundColor(paneSelectorHighlightColor)
	case SDIAG_PAGE:
		a.TabSchedulerBox.SetBackgroundColor(paneSelectorHighlightColor)
	case SACCTMGR_PAGE:
		a.TabAccountingBox.SetBackgroundColor(paneSelectorHighlightColor)
	}
}

func (a *App) ShowNodeDetails(nodeName string) {
	details, err := model.GetNodeDetailsWithTimeout(nodeName, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching node details:\n%s", err.Error())
	}
	a.ShowModalPopup(fmt.Sprintf("Node Details: %s", nodeName), details)
}

func (a *App) ShowJobDetails(jobID string) {
	details, err := model.GetJobDetailsWithTimeout(jobID, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching job details:\n%s", err.Error())
	}
	a.ShowModalPopup(fmt.Sprintf("Job Details: %s", jobID), details)
}
