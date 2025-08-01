package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	SORT_NONE = 0
	SORT_ASC  = 1
	SORT_DESC = -1
)

func (a *App) GetCurrentStuiView() *StuiView {
	switch a.CurrentTableView {
	case a.NodesView.Table:
		return a.NodesView
	case a.JobsView.Table:
		return a.JobsView
	case a.SacctView.Table:
		return a.SacctView
	case a.SacctMgrView.Table:
		return a.SacctMgrView
	default:
		return nil
	}
}

func (a *App) ClearSelectionFromCurrentView() {
	view := a.GetCurrentStuiView()
	if view != nil {
		view.Selection = map[string]bool{}
	}
}

func (a *App) GetProviderForPage(page string) model.DataProvider[*model.TableData] {
	switch page {
	case NODES_PAGE:
		return a.NodesProvider
	case JOBS_PAGE:
		return a.JobsProvider
	case SACCT_PAGE:
		return a.SacctProvider
	case SACCTMGR_PAGE:
		return a.SacctMgrProvider
	default:
		return nil
	}
}

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
	case SACCT_PAGE:
		if refresh {
			a.SacctProvider.Fetch()
		}
		a.SacctView.SetFilter(config.PartitionFilter)
		a.SacctView.Render()
	case SDIAG_PAGE:
		if refresh {
			d := a.SdiagProvider.Data()
			a.SchedView.SetText(d.Data)
		}
		// No rendering operation needed, TextView just gets its data set periodically
	}
	go a.App.QueueUpdateDraw(func() {})
}

func (a *App) ShowModalPopupTable(title string, table *tview.Table) {
	a.showModalPopup(title, table, 16, 10, 0)
}

func (a *App) ShowModalPopupString(title, details string) {
	detailView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true). // Enable text wrapping
		SetTextAlign(tview.AlignLeft)
	detailView.SetText(details)

	a.showModalPopup(title, detailView, 16, 10, 0)
}
func (a *App) ShowModalPopupMinimal(details string) {
	detailView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true). // Enable text wrapping
		SetTextAlign(tview.AlignLeft)
	detailView.SetText(details)

	a.showModalPopup("Full cell contents", detailView, 5, 10, 1)
}

func (a *App) showModalPopup(title string, primitive tview.Primitive, width int, height int, verticalPadding int) {
	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(fmt.Sprintf(" %s (ESC to close) ", title)).
			SetTextColor(generalTextColor),
			2, 0, false).
		AddItem(primitive, 0, 1, true)

	modal.SetBorder(true).
		SetBorderColor(modalBorderColor).
		SetBackgroundColor(generalBackgroundColor)

	// Create centered container with fixed size (80% width, 90% height)
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(modal, 0, verticalPadding, true).
			AddItem(modal, 0, height, true). // Increased height
			AddItem(nil, 0, verticalPadding, true),
			0, width, false). // Increased width
		AddItem(nil, 0, 1, false)

	// Store current page before showing modal
	previousFocus := a.App.GetFocus()

	// Add as overlay without switching pages
	time := time.Now().Unix()
	pageName := fmt.Sprintf("detailView-%d", time)
	a.Pages.AddPage(pageName, centered, true, true)
	a.App.SetFocus(primitive)

	// Set up handler to return to correct view when closed
	// Sadly we need to cast the primitive to an acceptable shape
	// in order to define input capture.
	table, is_table := primitive.(*tview.Table)
	textv, is_textv := primitive.(*tview.TextView)
	if is_textv {
		textv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				a.Pages.RemovePage(pageName)
				a.App.SetFocus(previousFocus)
				return nil
			}
			return event
		})
	}
	if is_table {
		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				a.Pages.RemovePage(pageName)
				a.App.SetFocus(previousFocus)
				return nil
			}
			return event
		})
	}
}

func (a *App) setActiveTab(active string) {
	// Reset all to black
	a.TabNodesBox.SetBackgroundColor(generalBackgroundColor)
	a.TabJobsBox.SetBackgroundColor(generalBackgroundColor)
	a.TabSchedulerBox.SetBackgroundColor(generalBackgroundColor)
	a.TabAccountingMgrBox.SetBackgroundColor(generalBackgroundColor)
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
		a.TabAccountingMgrBox.SetBackgroundColor(paneSelectorHighlightColor)
	case SACCT_PAGE:
		a.TabAccountingBox.SetBackgroundColor(paneSelectorHighlightColor)
	}
}

func (a *App) ShowNodeDetails(nodeName string) {
	details, err := model.GetNodeDetailsWithTimeout(nodeName, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching node details:\n%s", err.Error())
	}
	a.ShowModalPopupString(fmt.Sprintf("Node Details: %s", nodeName), details)
}

func (a *App) ShowJobDetails(jobID string) {
	details, err := model.GetJobDetailsWithTimeout(jobID, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching job details:\n%s", err.Error())
	}
	a.ShowModalPopupString(fmt.Sprintf("Job Details [scontrol]: %s", jobID), details)
}

func (a *App) ShowSacctJobDetails(jobID string) {
	details, err := model.GetSacctJobDetailsWithTimeout(jobID, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching job details:\n%s", err.Error())
	}

	// Process the parsable input (pipe-separated) into a nice table
	table := tview.NewTable()
	table.SetEvaluateAllRows(true) // Enable this so rows do not jump. Expensive but fine for a small table.
	table.SetFixed(1, 1)
	table.SetBorderPadding(0, 0, 1, 1)

	for i, line := range strings.Split(details, "\n") {
		for j, cell := range strings.Split(line, "|") {
			tc := tview.NewTableCell(cell)
			if i == 0 {
				tc.SetAlign(2) // Align the title column to the right
				tc.SetAttributes(tcell.AttrBold)
				tc.SetTextColor(selectionColor)
			}
			if j == 0 {
				tc.SetAttributes(tcell.AttrBold)
				tc.SetTextColor(selectionColor)
			}
			if i != 0 && j != 0 {
				tc.SetClickedFunc(func() bool {
					a.ShowModalPopupMinimal(cell)
					return true
				})
			}

			table.SetCell(j, i, tc)
		}
	}

	a.ShowModalPopupTable(fmt.Sprintf("Job Details [sacct]: %s", jobID), table)
}

// FormatNumberWithCommas adds thousand separators to an integer
// e.g., 1000 -> "1,000", 1000000 -> "1,000,000"
func FormatNumberWithCommas(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	// Start from the right and work backwards
	str := strconv.Itoa(n)
	result := ""
	for i := len(str) - 1; i >= 0; i-- {
		if (len(str)-i-1)%3 == 0 && i != len(str)-1 {
			result = "," + result
		}
		result = string(str[i]) + result
	}

	return sign + result
}

func PadSelectorTitle(title string) string {
	titleWidth := 17
	if len(title) < titleWidth {
		return fmt.Sprintf("%-*s", titleWidth, title)
	}
	return title
}
