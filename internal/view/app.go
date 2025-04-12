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
	App              *tview.Application
	Pages            *tview.Pages
	PagesContainer   *tview.Flex // Container for pages with border title
	LastUpdate       time.Time
	LastReqDuration  time.Duration
	startTime        time.Time    // Start time of the application
	CurrentTableView *tview.Table // Points to either NodesView or JobsView

	// Base app components
	HeaderGrid              *tview.Grid
	HeaderGridInnerContents *tview.Grid
	MainFlex                *tview.Flex
	FooterGrid              *tview.Grid

	// Main views and their grids
	NodesView *tview.Table
	JobsView  *tview.Table
	SchedView *tview.TextView
	AcctView  *tview.Table

	NodeGrid *tview.Grid
	JobGrid  *tview.Grid
	AcctGrid *tview.Grid

	// Selections for each view
	SelectedNodes    map[string]bool // Track selected nodes by name
	SelectedJobs     map[string]bool // Track selected jobs by ID
	SelectedAcctRows map[string]bool // Track selected acc rows

	// Footer
	HeaderLineOne *tview.TextView
	HeaderLineTwo *tview.TextView // Combined status line
	FooterMessage *tview.TextView

	// Current tab indicators
	TabNodesBox      *tview.TextView
	TabJobsBox       *tview.TextView
	TabSchedulerBox  *tview.TextView
	TabAccountingBox *tview.TextView

	// Partition selector
	PartitionSelector            *tview.DropDown
	PartitionSelectorFirstUpdate bool

	// SacctMgr entity selector
	SacctMgrEntitySelector            *tview.DropDown
	SacctMgrEntitySelectorFirstUpdate bool

	// Search state
	SearchBox     *tview.InputField
	SearchActive  bool
	SearchPattern string
	searchTimer   *time.Timer // Timer for debouncing search updates

	// Command modal state
	CommandModalOpen bool

	// Stored Data
	DataLoaded     chan struct{} // Channel to signal data has been loaded
	NodesTableData *model.TableData
	JobsTableData  *model.TableData
	AcctTableData  *model.TableData
	PartitionsData *model.TableData
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
		startTime:                         time.Now(),
		App:                               tview.NewApplication(),
		Pages:                             tview.NewPages(),
		DataLoaded:                        make(chan struct{}),
		PartitionSelectorFirstUpdate:      true,
		SacctMgrEntitySelectorFirstUpdate: true,
		HeaderGridInnerContents:           tview.NewGrid(),
	}

	// Init selectors, otherwise segfault lol
	application.SelectedNodes = make(map[string]bool)
	application.SelectedJobs = make(map[string]bool)
	application.SelectedAcctRows = make(map[string]bool)
	return &application
}

func (a *App) SetupViews() {
	a.SetupSearchBox()
	a.SetupPartitionSelector()

	if config.SacctEnabled {
		a.SetupSacctMgrEntitySelector()
	}

	// HeaderLineOne components
	a.FooterMessage = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	a.HeaderLineOne = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Combined status line
	a.HeaderLineTwo = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Create tab boxes
	a.TabNodesBox = tview.NewTextView().
		SetText("(1) Nodes")
	a.TabNodesBox.SetBackgroundColor(tcell.ColorDarkOrange)
	a.TabJobsBox = tview.NewTextView().
		SetText("(2) Jobs")
	a.TabSchedulerBox = tview.NewTextView().
		SetText("(3) Scheduler")
	a.TabAccountingBox = tview.NewTextView().
		SetText("(4) Acct Manager")

	// Create a grid for the tabs
	tabGrid := tview.NewGrid().
		// SetRows(1,1,1).
		AddItem(a.TabNodesBox, FRST_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabJobsBox, SCND_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabSchedulerBox, THRD_ROW, FRST_COL, 1, 1, 1, 0, false)

	if config.SacctEnabled {
		tabGrid.AddItem(a.TabAccountingBox, FRTH_ROW, FRST_COL, 1, 1, 1, 0, false)
	}

	a.HeaderGrid = tview.NewGrid().
		SetColumns(-1, -2, -1).
		SetBorders(true).
		AddItem(a.HeaderGridInnerContents, FRST_ROW, FRST_COL, 1, 1, 0, 0, false).
		AddItem(
			tview.NewGrid().
				SetRows(-1, -1).
				AddItem(a.HeaderLineOne, FRST_ROW, FRST_COL, 1, 1, 0, 0, false).
				AddItem(a.HeaderLineTwo, SCND_ROW, FRST_COL, 1, 1, 0, 0, false).
				AddItem(a.FooterMessage, THRD_ROW, FRST_COL, 1, 1, 0, 0, false),
			FRST_ROW, SCND_COL, 1, 1, 0, 0, false).
		AddItem(tabGrid, FRST_ROW, THRD_COL, 1, 1, 0, 0, false)

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
	a.MainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.HeaderGrid, 6, 0, false).
		AddItem(a.PagesContainer, 0, 1, true)

	a.MainFlex.SetBorder(true).
		SetBorderAttributes(tcell.AttrDim).
		SetTitle(" stui - Slurm Management TUI ").
		SetTitleAlign(tview.AlignCenter)

	// Nodes View
	a.NodesView = tview.NewTable()
	a.NodesView.
		SetBorders(false). // Remove all borders
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

	// Jobs View
	a.JobsView = tview.NewTable()
	a.JobsView.
		SetBorders(false). // Remove all borders
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

	// Accounting view
	if config.SacctEnabled {
		a.AcctView = tview.NewTable()
		a.AcctView.
			SetBorders(false). // Remove all borders
			SetTitleAlign(tview.AlignLeft).
			SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
		a.AcctView.SetFixed(1, 0)             // Fixed header row
		a.AcctView.SetSelectable(true, false) // Selectable rows but not columns
		// Configure more compact highlighting
		a.AcctView.SetSelectedStyle(tcell.StyleDefault.
			Background(rowCursorColorBackground).
			Foreground(rowCursorColorForeground))
		a.AcctView.SetBackgroundColor(tcell.ColorBlack) // Add this line
		a.AcctGrid = tview.NewGrid().
			SetRows(0). // Just table initially
			SetColumns(0).
			AddItem(a.AcctView, 0, 0, 1, 1, 0, 0, true)
		a.Pages.AddPage("accounting", a.AcctGrid, true, false)
	}

	// Scheduler View
	{
		a.SchedView = tview.NewTextView()
		a.SchedView.
			SetDynamicColors(true).
			SetScrollable(true).
			SetWrap(false).
			SetTitleAlign(tview.AlignLeft).
			SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
		a.Pages.AddPage("scheduler", a.SchedView, true, false)
	}

	{ // Starting position
		a.CurrentTableView = a.NodesView
		a.SetHeaderGridInnerContents(a.PartitionSelector)
	}
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

	// Nodes data
	a.NodesTableData, err = model.GetNodesWithTimeout(config.RequestTimeout)
	a.closeOnError(err)
	a.RenderTable(a.NodesView, *a.NodesTableData)

	// Jobs data
	a.JobsTableData, err = model.GetJobsWithTimeout(config.RequestTimeout)
	a.closeOnError(err)
	a.RenderTable(a.JobsView, *a.JobsTableData)

	// Sacctmgr data
	if config.SacctEnabled {
		_, entity := a.SacctMgrEntitySelector.GetCurrentOption()

		a.AcctTableData, err = model.GetSacctMgrEntityWithTimeout(entity, config.RequestTimeout)
		a.closeOnError(err) // TODO: This is lazy and won't work properly if user gets e.g. permission denied
		a.RenderTable(a.AcctView, *a.AcctTableData)
	}

	// Scheduler data
	sdiagOutput, err := model.GetSdiagWithTimeout(config.RequestTimeout)
	a.closeOnError(err)
	a.SchedView.SetText(sdiagOutput)

	a.LastReqDuration = time.Since(start)
	a.LastUpdate = time.Now()

	// Update status line immediately
	schedulerHost, schedulerIP := model.GetSchedulerInfoWithTimeout(config.RequestTimeout)
	a.UpdateHeader(schedulerHost, schedulerIP)

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
	case a.AcctView:
		a.RenderTable(table, *a.AcctTableData)
	default:
		return
	}
}

func (a *App) RenderTable(table *tview.Table, data model.TableData) {
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
		} else if table == a.AcctView {
			_, entity := a.SacctMgrEntitySelector.GetCurrentOption()
			a.PagesContainer.SetTitle(fmt.Sprintf(
				" %s rows (%d / %d) ",
				entity,
				filteredCount,
				totalCount,
			))
		}
	}

	// Set headers with fixed widths and padding
	for col, header := range *data.Headers {

		// If header is a divided type, clean it up
		headerName := header.Name
		if header.DividedByColumn {
			headerName = strings.Replace(header.Name, "//", "/", 1)
		}

		// Pad header with spaces to maintain width
		paddedHeader := fmt.Sprintf("%-*s", header.Width, headerName)
		table.SetCell(0, col, tview.NewTableCell(paddedHeader).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
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
		} else if table == a.AcctView {
			_, entity := a.SacctMgrEntitySelector.GetCurrentOption()
			a.PagesContainer.SetTitle(fmt.Sprintf(
				" %s rows (%d / %d) ",
				entity,
				filteredCount,
				totalCount,
			))
		}
	}

	// Set rows with text wrapping
	for row, rowData := range filteredRows {
		for col, cell := range rowData {
			cellView := tview.NewTableCell(cell).
				SetAlign(tview.AlignLeft).
				SetMaxWidth((*data.Headers)[col].Width).
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
		for col, header := range *data.Headers {
			spaces := strings.Repeat(" ", header.Width)
			table.SetCell(1, col, tview.NewTableCell(spaces).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(header.Width).
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
	a.TabAccountingBox.SetBackgroundColor(tcell.ColorBlack)

	// Set active to orange
	switch active {
	case "nodes":
		a.TabNodesBox.SetBackgroundColor(tcell.ColorDarkOrange)
	case "jobs":
		a.TabJobsBox.SetBackgroundColor(tcell.ColorDarkOrange)
	case "scheduler":
		a.TabSchedulerBox.SetBackgroundColor(tcell.ColorDarkOrange)
	case "accounting":
		a.TabAccountingBox.SetBackgroundColor(tcell.ColorDarkOrange)
	}
}

func (a *App) ShowNodeDetails(nodeName string) {
	details, err := model.GetNodeDetailsWithTimeout(nodeName, config.RequestTimeout)
	if err != nil {
		details = fmt.Sprintf("Error fetching node details:\n%s", err.Error())
	}
	a.ShowModalPopup(fmt.Sprintf("Node Details: %s", nodeName), details)
}
