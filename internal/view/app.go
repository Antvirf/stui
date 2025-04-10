package view

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	rowCursorColorBackground = tcell.ColorDarkSlateGray
	rowCursorColorForeground = tcell.ColorWhite
	selectionColor           = tcell.ColorDarkOrange // The orange color used for selections
	selectionHighlightColor  = tcell.ColorLightGoldenrodYellow
)

type App struct {
	App             *tview.Application
	Pages           *tview.Pages
	TabNodesBox     *tview.TextView
	TabJobsBox      *tview.TextView
	TabSchedulerBox *tview.TextView
	SelectedNodes   map[string]bool // Track selected nodes by name
	SelectedJobs    map[string]bool // Track selected jobs by ID
	PagesContainer  *tview.Flex     // Container for pages with border title
	NodesView       *tview.Table
	JobsView        *tview.Table
	SchedView       *tview.TextView
	MainGrid        *tview.Flex
	LastUpdate      time.Time
	LastReqDuration time.Duration
	startTime       time.Time // Start time of the application

	// Footer
	FooterLineOne *tview.TextView
	FooterLineTwo *tview.TextView // Combined status line
	FooterMessage *tview.TextView

	// Partition selector
	PartitionSelector            *tview.DropDown
	PartitionSelectorFirstUpdate bool

	// Search state
	SearchBox        *tview.InputField
	SearchActive     bool
	SearchPattern    string
	CurrentTableView *tview.Table // Points to either NodesView or JobsView
	NodeGrid         *tview.Grid  // Grid containing nodes view and search
	searchTimer      *time.Timer  // Timer for debouncing search updates
	JobGrid          *tview.Grid  // Grid containing jobs view and search

	// Command modal state
	CommandModalOpen bool

	// Stored Data
	DataLoaded     chan struct{} // Channel to signal data has been loaded
	NodesTableData *model.TableData
	JobsTableData  *model.TableData
	PartitionsData *model.TableData

	// Footer stats
	FooterGrid      *tview.Grid
	FooterNodeStats *tview.TextView
	FooterJobStats  *tview.TextView
}

// Exit and log error details
func (a *App) closeOnError(err error) {
	if err != nil {
		a.App.Stop()
		log.Fatal(err)
	}
}

// Initializes a `stui` instance tview Application using the config module
func InitializeApplication() (a *App) {
	application := App{
		App:                          tview.NewApplication(),
		Pages:                        tview.NewPages(),
		DataLoaded:                   make(chan struct{}),
		PartitionSelectorFirstUpdate: true,
		startTime:                    time.Now(),
	}

	// Init selectors
	application.SelectedNodes = make(map[string]bool)
	application.SelectedJobs = make(map[string]bool)
	return &application
}

func (a *App) SetupViews() {
	a.SetupSearchBox()
	a.SetupPartitionSelector()

	// FooterLineOne components
	a.FooterMessage = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	a.FooterLineOne = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Combined status line
	a.FooterLineTwo = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Left footer section
	footerLeft := tview.NewGrid().
		SetRows(-1, -1).
		AddItem(a.FooterMessage, 0, 0, 1, 1, 0, 0, false).
		AddItem(a.FooterLineOne, 1, 0, 1, 1, 0, 0, false).
		AddItem(a.FooterLineTwo, 2, 0, 1, 1, 0, 0, false)

	// Right footer section with stats
	a.FooterNodeStats = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	a.FooterJobStats = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)

	footerRight := tview.NewGrid().
		SetRows(-1, -1, -1).
		AddItem(tview.NewBox(), 0, 0, 1, 1, 0, 0, false).
		AddItem(a.FooterNodeStats, 1, 0, 1, 1, 0, 0, false).
		AddItem(a.FooterJobStats, 2, 0, 1, 1, 0, 0, false)

	// Create tab boxes
	a.TabNodesBox = tview.NewTextView().
		SetText("(1) Nodes")
	a.TabNodesBox.SetBackgroundColor(tcell.ColorDarkOrange)

	a.TabJobsBox = tview.NewTextView().
		SetText("(2) Jobs")

	a.TabSchedulerBox = tview.NewTextView().
		SetText("(3) Scheduler")

	// Create a grid for the tabs
	tabGrid := tview.NewGrid().
		// SetRows(1,1,1).
		AddItem(a.TabNodesBox, 0, 0, 1, 1, 1, 0, false).
		AddItem(a.TabJobsBox, 1, 0, 1, 1, 1, 0, false).
		AddItem(a.TabSchedulerBox, 2, 0, 1, 1, 1, 0, false)

	// Combined footer grid
	a.FooterGrid = tview.NewGrid().
		SetColumns(-1, -10, -10, -10, -1).
		AddItem(tview.NewBox(), 0, 0, 1, 1, 0, 0, false).
		AddItem(footerLeft, 0, 1, 1, 1, 0, 0, false).
		AddItem(tabGrid, 0, 2, 1, 1, 0, 0, false).
		AddItem(footerRight, 0, 3, 1, 1, 0, 0, false).
		AddItem(tview.NewBox(), 0, 4, 1, 1, 0, 0, false)

	a.FooterGrid.SetBorder(true).SetBorderStyle(
		tcell.StyleDefault.
			Foreground(tcell.ColorGray).
			Background(tcell.ColorBlack),
	)

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
		AddItem(a.PartitionSelector, 1, 0, false).
		AddItem(a.PagesContainer, 0, 1, true).
		AddItem(a.FooterGrid, 5, 0, false)

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
		Background(rowCursorColorBackground).
		Foreground(rowCursorColorForeground))
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
		Background(rowCursorColorBackground).
		Foreground(rowCursorColorForeground))
	a.JobsView.SetBackgroundColor(tcell.ColorBlack) // Add this line

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

	a.FooterLineTwo.SetText("[::i]Data as of never (0 ms) - updating in 3s[::-]")
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

	// Things to do only after first tick of data has loaded
	go func() {
		<-a.DataLoaded
		// Partition selector relies on partition data being available
		a.setupPartitionSelectorOptions()

		// Scroll to the beginning of tables once at the start
		a.NodesView.ScrollToBeginning()
		a.JobsView.ScrollToBeginning()
	}()
}

func (a *App) UpdateAllViews() {
	if a.App == nil || a.NodesView == nil {
		return
	}

	start := time.Now()
	var err error

	a.PartitionsData, err = model.GetAllPartitionsWithTimeout(config.RequestTimeout)
	a.closeOnError(err)

	a.NodesTableData, err = model.GetNodesWithTimeout(config.RequestTimeout)
	a.closeOnError(err)
	a.RenderTable(a.NodesView, *a.NodesTableData)

	// Update jobs view with squeue output
	a.JobsTableData, err = model.GetJobsWithTimeout(config.RequestTimeout)
	a.closeOnError(err)
	a.RenderTable(a.JobsView, *a.JobsTableData)

	// Update scheduler view with sdiag output
	sdiagOutput, err := model.GetSdiagWithTimeout(config.RequestTimeout)
	a.closeOnError(err)
	a.SchedView.SetText(sdiagOutput)

	a.LastReqDuration = time.Since(start)
	a.LastUpdate = time.Now()

	// Update status line immediately
	schedulerHost, schedulerIP := model.GetSchedulerInfoWithTimeout(config.RequestTimeout)
	a.UpdateFooter(schedulerHost, schedulerIP)

	// Inform that data has been loaded
	select {
	case a.DataLoaded <- struct{}{}:
	default:
	}
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
	columnWidths := []int{15, 15, 15, 20, 15, 10, 20, 6, 6, 6, 20} // Wider columns for Job Name, State and Nodes

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

func (a *App) ShowJobDetails(jobID string) {
	details, err := model.GetJobDetailsWithTimeout(jobID, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching job details:\n%s", err.Error())
	}
	a.ShowModalPopup(fmt.Sprintf("Job Details: %s", jobID), details)
}

func (a *App) setActiveTab(active string) {
	// Reset all to black
	a.TabNodesBox.SetBackgroundColor(tcell.ColorBlack)
	a.TabJobsBox.SetBackgroundColor(tcell.ColorBlack)
	a.TabSchedulerBox.SetBackgroundColor(tcell.ColorBlack)

	// Set active to orange
	switch active {
	case "nodes":
		a.TabNodesBox.SetBackgroundColor(tcell.ColorDarkOrange)
	case "jobs":
		a.TabJobsBox.SetBackgroundColor(tcell.ColorDarkOrange)
	case "scheduler":
		a.TabSchedulerBox.SetBackgroundColor(tcell.ColorDarkOrange)
	}
}

func (a *App) ShowNodeDetails(nodeName string) {
	details, err := model.GetNodeDetailsWithTimeout(nodeName, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching node details:\n%s", err.Error())
	}
	a.ShowModalPopup(fmt.Sprintf("Node Details: %s", nodeName), details)
}
