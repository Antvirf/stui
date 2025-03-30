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
	refreshInterval time.Duration
}

func main() {
	app := &App{
		app:            tview.NewApplication(),
		pages:          tview.NewPages(),
		refreshInterval: 5 * time.Second,
	}

	app.setupViews()
	app.setupKeybinds()
	app.startRefresh()

	if err := app.app.SetRoot(app.pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (a *App) setupViews() {
	// Nodes View
	a.nodesView = tview.NewTextView().
		SetDynamicColors(true).
		SetTitle(" Nodes (1) ").
		SetTitleAlign(tview.AlignLeft)
	a.pages.AddPage("nodes", a.nodesView, true, true)

	// Jobs View
	a.setupJobsView()
	a.pages.AddPage("jobs", a.jobsView, true, false)

	// Scheduler View
	a.schedView = tview.NewTextView().
		SetTitle(" Scheduler (3) ").
		SetTitleAlign(tview.AlignLeft)
	a.pages.AddPage("scheduler", a.schedView, true, false)
}

func (a *App) setupJobsView() {
	a.jobsView = tview.NewTable().
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
		case '2':
			a.pages.SwitchToPage("jobs")
		case '3':
			a.pages.SwitchToPage("scheduler")
		}
		return event
	})
}

func (a *App) startRefresh() {
	ticker := time.NewTicker(a.refreshInterval)
	go func() {
		for range ticker.C {
			a.updateAllViews()
		}
	}()
}

func (a *App) updateAllViews() {
	a.app.QueueUpdateDraw(func() {
		if nodes, err := fetchNodes(); err == nil {
			a.nodesView.SetText(nodes)
		}
		// TODO: Add jobs and scheduler updates
	})
}

func fetchNodes() (string, error) {
	cmd := exec.Command("sinfo", "-N", "-o%N %P %c %m %G %T")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("sinfo failed: %v", err)
	}
	return string(out), nil
}
