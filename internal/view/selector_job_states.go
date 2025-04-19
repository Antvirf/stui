package view

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupJobStateSelector() {
	a.JobStateSelector = tview.NewDropDown().
		SetLabel("(s) State: ").
		SetLabelStyle(tcell.StyleDefault.Foreground(dropdownForegroundColor)).
		SetListStyles(
			tcell.StyleDefault,
			tcell.StyleDefault.Background(selectionColor),
		).
		SetFieldWidth(20).
		SetFieldBackgroundColor(dropdownBackgroundColor).
		SetTextOptions("  ", "  ", "", "", "")

	a.JobStateSelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
			return nil
		}
		return event
	})

	for _, entity := range model.SCONTROL_JOB_STATES {
		a.JobStateSelector.AddOption(
			entity,
			a.applyJobStateSelector(entity),
		)
	}
	a.JobStateSelector.SetCurrentOption(0)
}

func (a *App) applyJobStateSelector(entity string) func() {
	return func() {
		config.JobStateCurrentChoice = entity
		if a.FirstRenderComplete {
			a.RenderCurrentView()
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
		}
	}
}
