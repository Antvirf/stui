package view

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupJobStateSelector(storedSelectorValue string) {
	a.JobStateSelector = tview.NewDropDown().
		SetLabel(PadSelectorTitle("(s) State:")).
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

	for index, entity := range model.SCONTROL_JOB_STATES {
		a.JobStateSelector.AddOption(
			entity,
			a.applyJobStateSelector(entity),
		)

		// By default, show all states - index 0
		if index == 0 {
			a.JobStateSelector.SetCurrentOption(0)
		}

		// If storedSelectorValue matches anything, use that.
		if entity == storedSelectorValue {
			a.JobStateSelector.SetCurrentOption(index)
		}
	}
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
