package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableData struct {
	Headers []string
	Rows    [][]string
}

type App struct {
	app             *tview.Application
	pages           *tview.Pages
	nodesView       *tview.Table
	jobsView        *tview.Table
	schedView       *tview.TextView
	footer          *tview.TextView
	footerStatus    *tview.TextView
	footerSeparator *tview.Box
	mainGrid        *tview.Grid
	refreshInterval time.Duration
	requestTimeout  time.Duration
	lastUpdate      time.Time
	nextUpdate      time.Time
	lastReqDuration time.Duration
	lastReqError    error
	debugMultiplier int // Number of times to multiply node entries for debugging

	// Search state
	searchBox        *tview.InputField
	searchActive     bool
	searchPattern    string
	currentTableView *tview.Table // Points to either nodesView or jobsView
	nodeGrid         *tview.Grid  // Grid containing nodes view and search
	jobGrid          *tview.Grid  // Grid containing jobs view and search
}

func main() {
	debugMultiplier := 3 // Default multiplier value for debugging
	app := &App{
		app:             tview.NewApplication(),
		pages:           tview.NewPages(),
		refreshInterval: 3 * time.Second,
		requestTimeout:  2 * time.Second, // Must be less than refreshInterval
		debugMultiplier: debugMultiplier,
	}

	app.setupViews()
	app.setupKeybinds()
	app.startRefresh()

	if err := app.app.SetRoot(app.mainGrid, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}

func (a *App) setupSearchBox() {
	a.searchBox = tview.NewInputField().
		SetLabel("  Regex search (case-insensitive): ").
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			a.searchPattern = strings.TrimSpace(text)
			a.searchActive = a.searchPattern != ""
			if a.currentTableView != nil {
				a.updateTableView(a.currentTableView)
			}
		})
	a.searchBox.SetBorder(false)

	// Set up input capture for search box
	a.searchBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.hideSearchBox()
			a.updateTableView(a.currentTableView)
			return nil
		case tcell.KeyEnter:
			if a.searchPattern == "" {
				a.hideSearchBox()
			} else {
				a.app.SetFocus(a.currentTableView)
			}
			return nil
		}
		return event
	})
}

func (a *App) showSearchBox() {
	if a.currentTableView == nil {
		return
	}

	// Get the appropriate grid
	grid := a.nodeGrid
	if a.currentTableView == a.jobsView {
		grid = a.jobGrid
	}

	// Clear and rebuild the grid with search box
	grid.Clear()
	grid.SetRows(1, 0) // 1 row for search, rest for table
	grid.AddItem(a.searchBox, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(a.currentTableView, 1, 0, 1, 1, 0, 0, false)

	// Set focus and ensure search box is visible
	a.app.SetFocus(a.searchBox)
	a.searchActive = true
}

func (a *App) hideSearchBox() {
	if a.currentTableView == nil {
		return
	}

	// Get the appropriate grid
	grid := a.nodeGrid
	if a.currentTableView == a.jobsView {
		grid = a.jobGrid
	}

	// Clear and rebuild grid without search box
	grid.Clear()
	grid.SetRows(0) // Just table
	grid.AddItem(a.currentTableView, 0, 0, 1, 1, 0, 0, true)

	// Reset search state
	a.searchBox.SetText("")
	a.searchActive = false
	a.app.SetFocus(a.currentTableView)
}

func (a *App) setupViews() {
	a.setupSearchBox()
	// Footer components
	a.footerStatus = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("Nodes (1) - Jobs (2) - Scheduler (3)")

	footerGrid := tview.NewGrid().
		SetRows(1, 1). // 1 for status, 1 for tabs
		SetColumns(0). // Single column
		AddItem(a.footerStatus, 0, 0, 1, 1, 0, 0, false).
		AddItem(a.footer, 1, 0, 1, 1, 0, 0, false)
	footerGrid.SetBorder(true).
		SetBorderPadding(0, 0, 1, 0)

	a.footerSeparator = tview.NewBox().
		SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetBorderStyle(tcell.StyleDefault.
			Foreground(tcell.ColorGray).
			Background(tcell.ColorBlack))

	// Main grid layout
	a.mainGrid = tview.NewGrid().
		SetRows(0, 1, 4). // 0 for pages (flexible), 1 for separator, 4 for footer
		SetColumns(0).    // Single column
		AddItem(a.pages, 0, 0, 1, 1, 0, 0, true).
		AddItem(a.footerSeparator, 1, 0, 1, 1, 0, 0, false).
		AddItem(footerGrid, 2, 0, 1, 1, 0, 0, false)

	a.mainGrid.SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetTitle(" S9S - Slurm Management TUI ").
		SetTitleAlign(tview.AlignCenter)

	// Nodes View
	a.nodesView = tview.NewTable()
	a.nodesView.
		SetBorders(false). // Remove all borders
		SetTitle(" Nodes (1) ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	a.nodesView.SetFixed(1, 0)             // Fixed header row
	a.nodesView.SetSelectable(true, false) // Selectable rows but not columns
	// Configure more compact highlighting
	a.nodesView.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkSlateGray).
		Foreground(tcell.ColorWhite))

	a.nodeGrid = tview.NewGrid().
		SetRows(0). // Just table initially
		SetColumns(0).
		AddItem(a.nodesView, 0, 0, 1, 1, 0, 0, true)
	a.pages.AddPage("nodes", a.nodeGrid, true, true)
	a.currentTableView = a.nodesView

	// Jobs View
	a.setupJobsView()
	a.jobGrid = tview.NewGrid().
		SetRows(0). // Just table initially
		SetColumns(0).
		AddItem(a.jobsView, 0, 0, 1, 1, 0, 0, true)
	a.pages.AddPage("jobs", a.jobGrid, true, false)

	// Scheduler View
	a.schedView = tview.NewTextView()
	a.schedView.
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false).
		SetTitle(" Scheduler (3) ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	a.pages.AddPage("scheduler", a.schedView, true, false)

	// Set initial active tab highlight and status
	a.footer.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
	a.footerStatus.SetText("[::i]Data as of never (0 ms) - updating in 3s[::-]")
}

func (a *App) setupJobsView() {
	a.jobsView = tview.NewTable()
	a.jobsView.
		SetBorders(false). // Remove all borders
		SetTitle(" Jobs (2) ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	a.jobsView.SetFixed(1, 0)             // Fixed header row
	a.jobsView.SetSelectable(true, false) // Selectable rows but not columns
	// Configure more compact highlighting
	a.jobsView.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorDarkSlateGray).
		Foreground(tcell.ColorWhite))

	headers := []string{"ID", "User", "Partition", "Name", "State", "Time", "Nodes"}
	for i, h := range headers {
		a.jobsView.SetCell(0, i, tview.NewTableCell(h).
			SetSelectable(false).
			SetAlign(tview.AlignCenter).
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite).
			SetAttributes(tcell.AttrBold))
	}
}

var appInstance *App

func GetApp() *App {
	return appInstance
}

func (a *App) setupKeybinds() {
	appInstance = a

	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			if a.searchActive || a.searchBox.HasFocus() {
				a.hideSearchBox()
				a.updateTableView(a.currentTableView)
				return nil
			}
		}

		switch event.Rune() {
		case '1':
			a.pages.SwitchToPage("nodes")
			a.footer.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
			a.currentTableView = a.nodesView
		case '2':
			a.pages.SwitchToPage("jobs")
			a.footer.SetText("Nodes (1) - [::b]Jobs (2)[::-] - Scheduler (3)")
			a.currentTableView = a.jobsView
		case '3':
			a.pages.SwitchToPage("scheduler")
			a.footer.SetText("Nodes (1) - Jobs (2) - [::b]Scheduler (3)[::-]")
			a.currentTableView = nil
		case '/':
			if a.currentTableView != nil {
				a.showSearchBox()
				a.updateTableView(a.currentTableView) // Refresh to show empty state
				return nil
			}
		}
		return event
	})

	a.searchBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.hideSearchBox()
			a.updateTableView(a.currentTableView)
			return nil
		case tcell.KeyEnter:
			if a.searchPattern == "" {
				a.hideSearchBox()
			} else {
				a.app.SetFocus(a.currentTableView)
			}
			return nil
		}
		return event
	})
}

func (a *App) startRefresh() {
	ticker := time.NewTicker(a.refreshInterval)
	go func() {
		for range ticker.C {
			a.app.QueueUpdateDraw(func() {
				a.updateAllViews()
			})
		}
	}()
	// Trigger initial update immediately
	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to let app start
		a.app.QueueUpdateDraw(func() {
			a.updateAllViews()
		})
	}()
}

func (a *App) updateAllViews() {
	if a.app == nil || a.nodesView == nil {
		return
	}

	start := time.Now()

	nodeData, err := a.fetchNodesWithTimeout()
	a.lastReqError = err
	if err == nil {
		RenderTable(a.nodesView, nodeData)
	}

	// Update jobs view with squeue output
	jobData, err := a.fetchJobsWithTimeout()
	if err == nil {
		RenderTable(a.jobsView, jobData)
	}

	// Update scheduler view with sdiag output
	sdiagOutput, err := a.fetchSdiagWithTimeout()
	if err == nil {
		a.schedView.SetText(sdiagOutput)
	}

	a.lastReqDuration = time.Since(start)
	a.lastUpdate = time.Now()
	a.nextUpdate = a.lastUpdate.Add(a.refreshInterval)

	// Update status footer immediately
	a.updateStatusFooter()

	// Start a ticker to update the countdown in real-time
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if time.Now().After(a.nextUpdate) {
				return
			}
			a.app.QueueUpdateDraw(a.updateStatusFooter)
		}
	}()

	// TODO: Add jobs and scheduler updates
}

func (a *App) updateTableView(table *tview.Table) {
	var data TableData
	switch table {
	case a.nodesView:
		data, _ = a.fetchNodesWithTimeout()
	case a.jobsView:
		data, _ = a.fetchJobsWithTimeout()
	default:
		return
	}

	table.Clear()
	a.renderTable(table, data)
}

func (a *App) renderTable(table *tview.Table, data TableData) {
	// Set headers with fixed width
	columnWidths := []int{10, 10, 10, 6, 8, 8, 20, 6, 6, 6, 15} // Adjust as needed

	// First clear the table but preserve column widths
	table.Clear()

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
	if a.searchActive {
		filteredRows = [][]string{}
		for _, row := range data.Rows {
			for _, cell := range row {
				if matched, _ := regexp.MatchString("(?i)"+a.searchPattern, cell); matched {
					filteredRows = append(filteredRows, row)
					break
				}
			}
		}
	}

	// Set rows with text wrapping
	for row, rowData := range filteredRows {
		for col, cell := range rowData {
			table.SetCell(row+1, col, tview.NewTableCell(cell).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(columnWidths[col]).
				SetExpansion(1))
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

func RenderTable(table *tview.Table, data TableData) {
	app := GetApp() // Need to add this function
	app.renderTable(table, data)
}

func (a *App) updateStatusFooter() {
	timeLeft := time.Until(a.nextUpdate).Round(time.Second)
	var status string
	if a.lastReqError != nil {
		status = fmt.Sprintf(
			"[::i]Data as of %s (FAILED) - updating in %s[::-]",
			a.lastUpdate.Format("15:04:05"),
			timeLeft,
		)
	} else {
		status = fmt.Sprintf(
			"[::i]Data as of %s (%d ms) - updating in %s[::-]",
			a.lastUpdate.Format("15:04:05"),
			a.lastReqDuration.Milliseconds(),
			timeLeft,
		)
	}
	a.footerStatus.SetText(status)
}

func (a *App) fetchJobsWithTimeout() (TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "squeue", "--noheader", "-o=%i|%u|%P|%j|%T|%M|%N")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return TableData{}, fmt.Errorf("timeout after %v", a.requestTimeout)
		}
		return TableData{}, fmt.Errorf("squeue failed: %v", err)
	}

	headers := []string{"ID", "User", "Partition", "Name", "State", "Time", "Nodes"}
	var rows [][]string

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) >= 7 {
			row := []string{
				strings.TrimPrefix(fields[0], "="), // Job ID
				strings.TrimPrefix(fields[1], "="), // User
				strings.TrimPrefix(fields[2], "="), // Partition
				strings.TrimPrefix(fields[3], "="), // Name
				strings.TrimPrefix(fields[4], "="), // State
				strings.TrimPrefix(fields[5], "="), // Time
				strings.TrimPrefix(fields[6], "="), // Nodes
			}
			rows = append(rows, row)
		}
	}

	return TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func (a *App) fetchSdiagWithTimeout() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sdiag")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", a.requestTimeout)
		}
		return "", fmt.Errorf("sdiag failed: %v", err)
	}

	return string(out), nil
}

func (a *App) fetchNodesWithTimeout() (TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sinfo", "--Node", "--noheader", "-o=%N|%P|%T|%c|%m|%L|%E|%f|%F|%G|%X|%Y|%Z")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return TableData{}, fmt.Errorf("timeout after %v", a.requestTimeout)
		}
		return TableData{}, fmt.Errorf("sinfo failed: %v", err)
	}

	headers := []string{
		"Node", "Partition", "State", "CPUs", "Memory",
		"CPULoad", "Reason", "Sockets", "Cores", "Threads", "GRES",
	}
	var rows [][]string

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) >= 11 {
			// Multiply the row according to debugMultiplier
			for i := 0; i < a.debugMultiplier; i++ {
				row := []string{
					strings.TrimPrefix(fields[0], "="), // Node
					fields[1],                          // Partition
					fields[2],                          // State
					fields[3],                          // CPUs
					fields[4],                          // Memory
					fields[5],                          // CPULoad
					fields[6],                          // Reason
					fields[7],                          // Sockets
					fields[8],                          // Cores
					fields[9],                          // Threads
					fields[10],                         // GRES
				}
				rows = append(rows, row)
			}
		}
	}

	return TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
}
