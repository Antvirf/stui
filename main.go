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
	flex           *tview.Flex
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

	if err := app.app.SetRoot(app.flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (a *App) setupViews() {
	// Footer with border
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 0).
		SetText("[::b][white:blue]Nodes (1)[::-] - [white:blue]Jobs (2)[::-] - [white:blue]Scheduler (3)[::-]")

	// Main layout with border
	a.flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.footer, 1, 1, false)
	a.flex.SetBorder(true)

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
	a.footer.SetText("[::b][green:blue]Nodes (1)[::-] - [white:blue]Jobs (2)[::-] - [white:blue]Scheduler (3)[::-]")
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
			a.footer.SetText("[::b][green:blue]Nodes (1)[::-] - [white:blue]Jobs (2)[::-] - [white:blue]Scheduler (3)[::-]")
		case '2':
			a.pages.SwitchToPage("jobs")
			a.footer.SetText("[::b][white:blue]Nodes (1)[::-] - [green:blue]Jobs (2)[::-] - [white:blue]Scheduler (3)[::-]")
		case '3':
			a.pages.SwitchToPage("scheduler")
			a.footer.SetText("[::b][white:blue]Nodes (1)[::-] - [white:blue]Jobs (2)[::-] - [green:blue]Scheduler (3)[::-]")
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
