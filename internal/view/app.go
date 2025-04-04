package view

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	selectionColor = tcell.ColorDarkOrange // The orange color used for selections
)

type App struct {
	App             *tview.Application
	Pages           *tview.Pages
	SelectedNodes   map[string]bool // Track selected nodes by name
	SelectedJobs    map[string]bool // Track selected jobs by ID
	PagesContainer  *tview.Flex     // Container for pages with border title
	NodesView       *tview.Table
	JobsView        *tview.Table
	SchedView       *tview.TextView
	Footer          *tview.TextView
	FooterStatus    *tview.TextView
	StatusLine      *tview.TextView // Combined status line
	FooterSeparator *tview.Box
	MainGrid        *tview.Flex
	LastUpdate      time.Time
	LastReqDuration time.Duration
	LastReqError    error
	// Config options, to be removed from here
	// DebugMultiplier        int // Number of times to multiply node entries for debugging
	// SearchDebounceInterval time.Duration
	// RequestTimeout         time.Duration

	// Search state
	SearchBox        *tview.InputField
	SearchActive     bool
	SearchPattern    string
	CurrentTableView *tview.Table // Points to either NodesView or JobsView
	NodeGrid         *tview.Grid  // Grid containing nodes view and search
	searchTimer      *time.Timer  // Timer for debouncing search updates
	JobGrid          *tview.Grid  // Grid containing jobs view and search

	// Stored Data
	NodesTableData *model.TableData
	JobsTableData  *model.TableData
}

// Initializes a `stui` instance tview Application using the config module
func InitializeApplication() (a *App) {
	application := App{
		App:   tview.NewApplication(),
		Pages: tview.NewPages(),
	}

	// Init selectors
	application.SelectedNodes = make(map[string]bool)
	application.SelectedJobs = make(map[string]bool)
	return &application
}

func (a *App) SetupViews() {
	a.SetupSearchBox()

	// Footer components
	a.FooterStatus = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	a.Footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("Nodes (1) - Jobs (2) - Scheduler (3)")

	// Combined status line
	a.StatusLine = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Parse slurm config to get scheduler info
	schedulerHost, schedulerIP := model.GetSchedulerInfo()
	a.UpdateStatusLine(a.StatusLine, schedulerHost, schedulerIP)

	footerGrid := tview.NewGrid().
		AddItem(a.Footer, 0, 0, 1, 1, 0, 0, false).
		AddItem(a.StatusLine, 1, 0, 1, 1, 0, 0, false)

	footerGrid.SetBorder(true).SetBorderStyle(
		tcell.StyleDefault.
			Foreground(tcell.ColorGray).
			Background(tcell.ColorBlack),
	).
		SetBorderPadding(0, 0, 0, 0)

	a.FooterSeparator = tview.NewBox().
		SetBorder(false).
		SetBorderAttributes(tcell.AttrBold)

	a.PagesContainer = tview.NewFlex().SetDirection(tview.FlexRow)

	a.PagesContainer.AddItem(a.Pages, 0, 30, true).
		SetBorder(true).
		SetBorderStyle(
			tcell.StyleDefault.
				Foreground(tcell.ColorGray).
				Background(tcell.ColorBlack),
		).
		SetTitle(" Nodes (0 / 0)") // Initial title matching nodes view

	// Main grid layout, implemented with Flex
	a.MainGrid = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.PagesContainer, 0, 30, true).
		AddItem(a.FooterSeparator, 0, 1, false).
		AddItem(footerGrid, 0, 3, false)

	a.MainGrid.SetBorder(true).
		SetBorderAttributes(tcell.AttrDim).
		SetTitle(" stui - Slurm Management TUI ").
		SetTitleAlign(tview.AlignCenter)

	// Nodes View
	a.NodesView = tview.NewTable()
	a.NodesView.
		SetBorders(false). // Remove all borders
		SetTitle(" Nodes (1) ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	a.NodesView.SetFixed(1, 0)             // Fixed header row
	a.NodesView.SetSelectable(true, false) // Selectable rows but not columns
	// Configure more compact highlighting
	a.NodesView.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkSlateGray).
		Foreground(tcell.ColorWhite))
	a.NodesView.SetBackgroundColor(tcell.ColorBlack) // Add this line

	a.NodeGrid = tview.NewGrid().
		SetRows(0). // Just table initially
		SetColumns(0).
		AddItem(a.NodesView, 0, 0, 1, 1, 0, 0, true)
	a.Pages.AddPage("nodes", a.NodeGrid, true, true)
	a.CurrentTableView = a.NodesView

	// Jobs View
	a.JobsView = tview.NewTable()
	a.JobsView.
		SetBorders(false). // Remove all borders
		SetTitle(" Jobs (2) ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	a.JobsView.SetFixed(1, 0)             // Fixed header row
	a.JobsView.SetSelectable(true, false) // Selectable rows but not columns
	// Configure more compact highlighting
	a.JobsView.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkSlateGray).
		Foreground(tcell.ColorWhite))
	a.JobsView.SetBackgroundColor(tcell.ColorBlack) // Add this line

	headers := []string{"ID", "User", "Partition", "Name", "State", "Time", "Nodes"}
	for i, h := range headers {
		a.JobsView.SetCell(0, i, tview.NewTableCell(h).
			SetSelectable(false).
			SetAlign(tview.AlignCenter).
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite).
			SetAttributes(tcell.AttrBold))
	}
	a.JobGrid = tview.NewGrid().
		SetRows(0). // Just table initially
		SetColumns(0).
		AddItem(a.JobsView, 0, 0, 1, 1, 0, 0, true)
	a.Pages.AddPage("jobs", a.JobGrid, true, false)

	// Scheduler View
	a.SchedView = tview.NewTextView()
	a.SchedView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false).
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	a.Pages.AddPage("scheduler", a.SchedView, true, false)

	// Set initial active tab highlight and status
	a.Footer.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
	a.FooterStatus.SetText("[::i]Data as of never (0 ms) - updating in 3s[::-]")
}

func (a *App) StartRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		// This fires a tick immediately, and then on an interval afterwards.
		for ; true; <-ticker.C {
			a.App.QueueUpdateDraw(func() {
				a.UpdateAllViews()
			})
		}
	}()
}

func (a *App) UpdateAllViews() {
	if a.App == nil || a.NodesView == nil {
		return
	}

	start := time.Now()
	var err error
	a.NodesTableData, err = model.GetNodesWithTimeout(config.RequestTimeout, config.DebugMultiplier)

	a.LastReqError = err
	if err == nil {
		a.RenderTable(a.NodesView, *a.NodesTableData)
	}
	// TODO? Panic?

	// Update jobs view with squeue output
	a.JobsTableData, err = model.GetJobsWithTimeout(config.RequestTimeout, config.DebugMultiplier)
	if err == nil {
		a.RenderTable(a.JobsView, *a.JobsTableData)
	}

	// Update scheduler view with sdiag output
	sdiagOutput, err := model.GetSdiagWithTimeout(config.RequestTimeout)
	if err == nil {
		a.SchedView.SetText(sdiagOutput)
	}

	a.LastReqDuration = time.Since(start)
	a.LastUpdate = time.Now()

	// Update status line immediately
	schedulerHost, schedulerIP := model.GetSchedulerInfo()
	a.UpdateStatusLine(a.StatusLine, schedulerHost, schedulerIP)
}

func (a *App) RerenderTableView(table *tview.Table) {
	table.Clear()
	switch table {
	case a.NodesView:
		a.RenderTable(table, *a.NodesTableData)
	case a.JobsView:
		a.RenderTable(table, *a.JobsTableData)
	default:
		return
	}

}

func (a *App) RenderTable(table *tview.Table, data model.TableData) {
	// Set headers with fixed width
	columnWidths := []int{10, 10, 10, 6, 8, 8, 20, 6, 6, 6, 15} // Adjust as needed

	// First clear the table but preserve column widths
	table.Clear()

	// Update page title with counts
	totalCount := len(data.Rows)
	filteredCount := totalCount
	if a.SearchActive {
		filteredCount = 0 // Will be updated in the filtering loop below
	}

	// Only update title if this is the currently active view
	if table == a.CurrentTableView {
		if table == a.NodesView {
			a.PagesContainer.SetTitle(fmt.Sprintf(" Nodes (%d / %d) ", filteredCount, totalCount))
		} else if table == a.JobsView {
			a.PagesContainer.SetTitle(fmt.Sprintf(" Jobs (%d / %d) ", filteredCount, totalCount))
		}
	}

	// Set headers with fixed widths and padding
	for col, header := range data.Headers {
		// Pad header with spaces to maintain width
		paddedHeader := fmt.Sprintf("%-*s", columnWidths[col], header)
		table.SetCell(0, col, tview.NewTableCell(paddedHeader).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetMaxWidth(columnWidths[col]).
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite).
			SetAttributes(tcell.AttrBold))
	}

	// Filter rows if search is active
	filteredRows := data.Rows
	if a.SearchActive {
		filteredRows = [][]string{}
		for _, row := range data.Rows {
			// Combine the entire row into a single string for regex matching
			rowString := strings.Join(row, " ")
			if matched, _ := regexp.MatchString("(?i)"+a.SearchPattern, rowString); matched {
				filteredRows = append(filteredRows, row)
				filteredCount++
			}
		}
		// Update title with filtered count
		if table == a.NodesView {
			a.PagesContainer.SetTitle(fmt.Sprintf(" Nodes (%d / %d) ", filteredCount, totalCount))
		} else if table == a.JobsView {
			a.PagesContainer.SetTitle(fmt.Sprintf(" Jobs (%d / %d) ", filteredCount, totalCount))
		}
	}

	// Set rows with text wrapping
	for row, rowData := range filteredRows {
		for col, cell := range rowData {
			cellView := tview.NewTableCell(cell).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(columnWidths[col]).
				SetExpansion(1)

			// Highlight selected rows
			if table == a.NodesView && a.SelectedNodes[rowData[0]] {
				cellView.SetBackgroundColor(selectionColor)
			} else if table == a.JobsView && a.SelectedJobs[rowData[0]] {
				cellView.SetBackgroundColor(selectionColor)
			} else {
				cellView.SetBackgroundColor(tcell.ColorBlack) // Explicitly set default when not selected
			}

			table.SetCell(row+1, col, cellView)
		}
	}

	// If no rows, set empty cells with spaces to maintain column widths
	if len(filteredRows) == 0 {
		for col, width := range columnWidths {
			// Create a cell with spaces to maintain width
			spaces := strings.Repeat(" ", width)
			table.SetCell(1, col, tview.NewTableCell(spaces).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(width).
				SetExpansion(1))
		}
	}
}

func (a *App) UpdateStatusLine(StatusLine *tview.TextView, host, ip string) {
	var status string
	if a.LastReqError != nil {
		status = fmt.Sprintf(
			"Scheduler: %s (%s) | Data as of %s (FAILED)",
			host,
			ip,
			a.LastUpdate.Format("15:04:05"),
		)
	} else {
		status = fmt.Sprintf(
			"Scheduler: %s (%s) | Data as of %s (%d ms)",
			host,
			ip,
			a.LastUpdate.Format("15:04:05"),
			a.LastReqDuration.Milliseconds(),
		)
	}
	StatusLine.SetText(status)
}

func (a *App) ShowDetailsModal(title, details string) {
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
			SetTextColor(tcell.ColorWhite),
			2, 0, false).
		AddItem(detailView, 0, 1, true)

	modal.SetBorder(true).
		SetBorderColor(tcell.ColorDarkOrange).
		SetBackgroundColor(tcell.ColorBlack)

	// Create centered container with fixed size (80% width, 90% height)
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(modal, 40, 10, true), // Increased height
			0, 16, false). // Increased width
		AddItem(nil, 0, 1, false)

	// Store current page before showing modal
	currentPage := "nodes"
	if a.CurrentTableView == a.JobsView {
		currentPage = "jobs"
	}

	// Add as overlay without switching pages
	pageName := "detailView"
	a.Pages.AddPage(pageName, centered, true, true)
	a.App.SetFocus(detailView)

	// Set up handler to return to correct view when closed
	detailView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			a.Pages.RemovePage(pageName)
			a.Pages.SwitchToPage(currentPage)
			a.App.SetFocus(a.CurrentTableView)
			return nil
		}
		return event
	})
}

func (a *App) ShowNodeDetails(nodeName string) {
	details, err := model.GetNodeDetailsWithTimeout(nodeName, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching node details:\n%s", err.Error())
	}
	a.ShowDetailsModal(fmt.Sprintf("Node Details: %s", nodeName), details)
}

func (a *App) ShowJobDetails(jobID string) {
	details, err := model.GetJobDetailsWithTimeout(jobID, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching job details:\n%s", err.Error())
	}
	a.ShowDetailsModal(fmt.Sprintf("Job Details: %s", jobID), details)
}
