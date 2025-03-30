package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app            *tview.Application
	pages          *tview.Pages
	nodesView      *tview.TextView
	jobsView       *tview.Table
	schedView      *tview.TextView
	footer         *tview.TextView
	footerSeparator *tview.Box
	mainGrid       *tview.Grid
	refreshInterval time.Duration
}

func main() {
	app := &App{
		app:            tview.NewApplication(),
		pages:          tview.NewPages(),
		refreshInterval: 3 * time.Second,
	}

	app.setupViews()
	app.setupKeybinds()
	app.startRefresh()

	if err := app.app.SetRoot(app.mainGrid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (a *App) setupViews() {
	// Footer components
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("Nodes (1) - Jobs (2) - Scheduler (3)").
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 0)

	a.footerSeparator = tview.NewBox().
		SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetBorderStyle(tcell.StyleDefault.
			Foreground(tcell.ColorGray).
			Background(tcell.ColorDefault))

	// Main grid layout
	a.mainGrid = tview.NewGrid().
		SetRows(0, 1, 3). // 0 for pages (flexible), 1 for separator, 3 for footer (with border)
		SetColumns(0).    // Single column
		AddItem(a.pages, 0, 0, 1, 1, 0, 0, true).
		AddItem(a.footerSeparator, 1, 0, 1, 1, 0, 0, false).
		AddItem(a.footer, 2, 0, 1, 1, 0, 0, false)
	
	a.mainGrid.SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetTitle(" S9S - Slurm Management TUI ").
		SetTitleAlign(tview.AlignCenter)

	// Nodes View
	a.nodesView = tview.NewTextView()
	a.nodesView.
		SetDynamicColors(true).
		SetTitle(" Nodes (1) ").
		SetTitleAlign(tview.AlignLeft)
	a.pages.AddPage("nodes", a.nodesView, true, true)

	// Jobs View
	a.setupJobsView()
	a.pages.AddPage("jobs", a.jobsView, true, false)

	// Scheduler View
	a.schedView = tview.NewTextView()
	a.schedView.
		SetTitle(" Scheduler (3) ").
		SetTitleAlign(tview.AlignLeft)
	a.pages.AddPage("scheduler", a.schedView, true, false)
	
	// Set initial active tab highlight
	a.footer.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
}

func (a *App) setupJobsView() {
	a.jobsView = tview.NewTable()
	a.jobsView.
		SetBorders(true).
		SetTitle(" Jobs (2) ").
		SetTitleAlign(tview.AlignLeft)

	headers := []string{"ID", "User", "Partition", "Name", "State", "Time", "Nodes"}
	for i, h := range headers {
		a.jobsView.SetCell(0, i, tview.NewTableCell(h).
			SetSelectable(false).
			SetAlign(tview.AlignCenter))
	}
}

func (a *App) setupKeybinds() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '1':
			a.pages.SwitchToPage("nodes")
			a.footer.SetText("[::b]Nodes (1)[::-] - Jobs (2) - Scheduler (3)")
		case '2':
			a.pages.SwitchToPage("jobs")
			a.footer.SetText("Nodes (1) - [::b]Jobs (2)[::-] - Scheduler (3)")
		case '3':
			a.pages.SwitchToPage("scheduler")
			a.footer.SetText("Nodes (1) - Jobs (2) - [::b]Scheduler (3)[::-]")
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
	
	if nodes, err := fetchNodes(); err == nil {
		a.nodesView.SetText(nodes)
	}
	// TODO: Add jobs and scheduler updates
}

func fetchNodes() (string, error) {
	cmd := exec.Command("sinfo", "-N", "-o%N %P %c %m %G %T")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("sinfo failed: %v", err)
	}
	return string(out), nil
}
