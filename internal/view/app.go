package view

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	App             *tview.Application
	Pages           *tview.Pages
	NodesView       *tview.Table
	JobsView        *tview.Table
	SchedView       *tview.TextView
	Footer          *tview.TextView
	FooterStatus    *tview.TextView
	StatusLine      *tview.TextView // Combined status line
	FooterSeparator *tview.Box
	MainGrid        *tview.Flex
	RefreshInterval time.Duration
	RequestTimeout  time.Duration
	LastUpdate      time.Time
	NextUpdate      time.Time
	LastReqDuration time.Duration
	LastReqError    error
	DebugMultiplier int // Number of times to multiply node entries for debugging

	// Search state
	SearchBox        *tview.InputField
	SearchActive     bool
	SearchPattern    string
	CurrentTableView *tview.Table // Points to either NodesView or JobsView
	NodeGrid         *tview.Grid  // Grid containing nodes view and search
	JobGrid          *tview.Grid  // Grid containing jobs view and search
}

var appInstance *App

func GetApp() *App {
	return appInstance
}

func (a *App) SetupSearchBox() {
	a.SearchBox = tview.NewInputField().
		SetLabel("  Regex search (case-insensitive): ").
		SetLabelColor(tcell.ColorDarkOrange).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {
			a.SearchPattern = strings.TrimSpace(text)
			wasActive := a.SearchActive
			a.SearchActive = a.SearchPattern != ""

			// Hide if search was cleared
			if wasActive && !a.SearchActive {
				a.HideSearchBox()
			}

			if a.CurrentTableView != nil {
				a.UpdateTableView(a.CurrentTableView)
			}
		})
	a.SearchBox.SetBorder(false)

	// Set up input capture for search box
	a.SearchBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.HideSearchBox()
			a.UpdateTableView(a.CurrentTableView)
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

func (a *App) ShowSearchBox() {
	if a.CurrentTableView == nil {
		return
	}

	// Get the appropriate grid
	grid := a.NodeGrid
	if a.CurrentTableView == a.JobsView {
		grid = a.JobGrid
	}

	// Clear and rebuild the grid with search box
	grid.Clear()
	grid.SetRows(1, 0)                                       // 1 row for search, rest for table
	grid.AddItem(a.SearchBox, 0, 0, 1, 1, 0, 0, false)       // Don't focus by default
	grid.AddItem(a.CurrentTableView, 1, 0, 1, 1, 0, 0, true) // Keep table focused

	a.SearchActive = true
}

func (a *App) HideSearchBox() {
	if a.CurrentTableView == nil {
		return
	}

	// Get the appropriate grid
	grid := a.NodeGrid
	if a.CurrentTableView == a.JobsView {
		grid = a.JobGrid
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
	schedulerHost, schedulerIP := a.GetSchedulerInfo()
	a.UpdateStatusLine(a.StatusLine, schedulerHost, schedulerIP)

	footerGrid := tview.NewGrid().
		AddItem(a.Footer, 0, 0, 1, 1, 0, 0, false).
		AddItem(a.StatusLine, 1, 0, 1, 1, 0, 0, false)

	footerGrid.SetBorder(true).
		SetBorderPadding(0, 0, 0, 0)

	a.FooterSeparator = tview.NewBox().
		SetBorder(false).
		SetBorderAttributes(tcell.AttrBold)

	// Main grid layout, implemented with Flex
	a.MainGrid = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.Pages, 0, 30, true).
		AddItem(a.FooterSeparator, 0, 1, false).
		AddItem(footerGrid, 0, 3, false)

	a.MainGrid.SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
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

	a.NodeGrid = tview.NewGrid().
		SetRows(0). // Just table initially
		SetColumns(0).
		AddItem(a.NodesView, 0, 0, 1, 1, 0, 0, true)
	a.Pages.AddPage("nodes", a.NodeGrid, true, true)
	a.CurrentTableView = a.NodesView

	// Jobs View
	a.SetupJobsView()
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
		SetTitle(" Scheduler (3) ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	a.Pages.AddPage("scheduler", a.SchedView, true, false)

	// Set initial active tab highlight and status
	a.Footer.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
	a.FooterStatus.SetText("[::i]Data as of never (0 ms) - updating in 3s[::-]")
}

func (a *App) SetupJobsView() {
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

	headers := []string{"ID", "User", "Partition", "Name", "State", "Time", "Nodes"}
	for i, h := range headers {
		a.JobsView.SetCell(0, i, tview.NewTableCell(h).
			SetSelectable(false).
			SetAlign(tview.AlignCenter).
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite).
			SetAttributes(tcell.AttrBold))
	}
}

func (a *App) StartRefresh() {
	ticker := time.NewTicker(a.RefreshInterval)
	go func() {
		for range ticker.C {
			a.App.QueueUpdateDraw(func() {
				a.UpdateAllViews()
			})
		}
	}()
	// Trigger initial update immediately
	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to let app start
		a.App.QueueUpdateDraw(func() {
			a.UpdateAllViews()
		})
	}()
}

func (a *App) UpdateAllViews() {
	if a.App == nil || a.NodesView == nil {
		return
	}

	start := time.Now()

	nodeData, err := a.FetchNodesWithTimeout()
	a.LastReqError = err
	if err == nil {
		RenderTable(a.NodesView, nodeData)
	}

	// Update jobs view with squeue output
	jobData, err := a.FetchJobsWithTimeout()
	if err == nil {
		RenderTable(a.JobsView, jobData)
	}

	// Update scheduler view with sdiag output
	sdiagOutput, err := a.FetchSdiagWithTimeout()
	if err == nil {
		a.SchedView.SetText(sdiagOutput)
	}

	a.LastReqDuration = time.Since(start)
	a.LastUpdate = time.Now()
	a.NextUpdate = a.LastUpdate.Add(a.RefreshInterval)

	// Update status line immediately
	schedulerHost, schedulerIP := a.GetSchedulerInfo()
	a.UpdateStatusLine(a.StatusLine, schedulerHost, schedulerIP)

	// TODO: Add jobs and scheduler updates
}

func (a *App) UpdateTableView(table *tview.Table) {
	var data model.TableData
	switch table {
	case a.NodesView:
		data, _ = a.FetchNodesWithTimeout()
	case a.JobsView:
		data, _ = a.FetchJobsWithTimeout()
	default:
		return
	}

	table.Clear()
	a.RenderTable(table, data)
}

func (a *App) RenderTable(table *tview.Table, data model.TableData) {
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
	if a.SearchActive {
		filteredRows = [][]string{}
		for _, row := range data.Rows {
			for _, cell := range row {
				if matched, _ := regexp.MatchString("(?i)"+a.SearchPattern, cell); matched {
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

func RenderTable(table *tview.Table, data model.TableData) {
	app := GetApp() // Need to add this function
	app.RenderTable(table, data)
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

func (a *App) FetchJobsWithTimeout() (model.TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.RequestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "squeue", "--noheader", "-o=%i|%u|%P|%j|%T|%M|%N")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return model.TableData{}, fmt.Errorf("timeout after %v", a.RequestTimeout)
		}
		return model.TableData{}, fmt.Errorf("squeue failed: %v", err)
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

	return model.TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
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
	details, err := a.FetchNodeDetailsWithTimeout(nodeName)
	if err != nil {
		details = fmt.Sprintf("Error fetching node details:\n%s", err.Error())
	}
	a.ShowDetailsModal(fmt.Sprintf("Node Details: %s", nodeName), details)
}

func (a *App) ShowJobDetails(jobID string) {
	details, err := a.FetchJobDetailsWithTimeout(jobID)
	if err != nil {
		details = fmt.Sprintf("Error fetching job details:\n%s", err.Error())
	}
	a.ShowDetailsModal(fmt.Sprintf("Job Details: %s", jobID), details)
}

func (a *App) FetchNodeDetailsWithTimeout(nodeName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.RequestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "scontrol", "show", "node", nodeName)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", a.RequestTimeout)
		}
		return "", fmt.Errorf("scontrol failed: %v", err)
	}
	return string(out), nil
}

func (a *App) FetchJobDetailsWithTimeout(jobID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.RequestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "scontrol", "show", "job", jobID)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", a.RequestTimeout)
		}
		return "", fmt.Errorf("scontrol failed: %v", err)
	}
	return string(out), nil
}

func (a *App) GetSchedulerInfo() (string, string) {
	// Get scheduler host from slurm config
	cmd := exec.Command("scontrol", "show", "config")
	out, err := cmd.Output()
	if err != nil {
		return "unknown", "unknown"
	}

	// Parse output for controller host
	var host string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "SlurmctldHost") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				// Extract host from SlurmctldHost[0]=hostname
				host = strings.TrimSpace(parts[1])
				if strings.Contains(host, "[") {
					host = strings.Split(host, "[")[0]
				}
				break
			}
		}
	}

	if host == "" {
		return "unknown", "unknown"
	}

	// Try to get IP
	addrs, err := net.LookupHost(host)
	if err == nil && len(addrs) > 0 {
		return host, addrs[0]
	}

	// Try short hostname if FQDN failed
	if strings.Contains(host, ".") {
		shortHost := strings.Split(host, ".")[0]
		addrs, err = net.LookupHost(shortHost)
		if err == nil && len(addrs) > 0 {
			return host, addrs[0]
		}
	}

	return host, "unknown"
}

func (a *App) FetchSdiagWithTimeout() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.RequestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sdiag")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", a.RequestTimeout)
		}
		return "", fmt.Errorf("sdiag failed: %v", err)
	}

	return string(out), nil
}

func (a *App) FetchNodesWithTimeout() (model.TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.RequestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sinfo", "--Node", "--noheader", "-o=%N|%P|%T|%c|%m|%L|%E|%f|%F|%G|%X|%Y|%Z")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return model.TableData{}, fmt.Errorf("timeout after %v", a.RequestTimeout)
		}
		return model.TableData{}, fmt.Errorf("sinfo failed: %v", err)
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
			// Multiply the row according to DebugMultiplier
			for i := 0; i < a.DebugMultiplier; i++ {
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

	return model.TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
}
