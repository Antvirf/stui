package view

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupNodeStateSelector(storedSelectorValue string) {
	a.NodeStateSelector = tview.NewDropDown().
		SetLabel(PadSelectorTitle("(s) State:")).
		SetLabelStyle(tcell.StyleDefault.Foreground(dropdownForegroundColor)).
		SetListStyles(
			tcell.StyleDefault,
			tcell.StyleDefault.Background(selectionColor),
		).
		SetFieldWidth(20).
		SetFieldBackgroundColor(dropdownBackgroundColor).
		SetTextOptions("  ", "  ", "", "", "")

	a.NodeStateSelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
			return nil
		}
		return event
	})

	for index, entity := range model.SCONTROL_NODE_STATES {
		a.NodeStateSelector.AddOption(
			entity,
			a.applyNodeStateSelector(entity),
		)

		// By default, show all states - index 0
		if index == 0 {
			a.NodeStateSelector.SetCurrentOption(0)
		}

		// If storedSelectorValue matches anything, use that.
		if entity == storedSelectorValue {
			a.NodeStateSelector.SetCurrentOption(index)
		}
	}
}

func (a *App) applyNodeStateSelector(entity string) func() {
	return func() {
		config.NodeStateCurrentChoice = entity
		if a.FirstRenderComplete {
			a.RenderCurrentView()
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
		}
	}
}
