package view

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupNodeStateSelector() {
	a.NodeStateSelector = tview.NewDropDown().
		SetLabel("(s) State: ").
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
			a.App.SetFocus(a.CurrentTableView)
			return nil
		}
		return event
	})

	for _, entity := range model.SCONTROL_NODE_STATES {
		a.NodeStateSelector.AddOption(
			entity,
			a.applyNodeStateSelector(entity),
		)
	}
	a.NodeStateSelector.SetCurrentOption(0)
}

func (a *App) applyNodeStateSelector(entity string) func() {
	return func() {
		config.NodeStateCurrentChoice = entity
		if a.FirstRenderComplete {
			a.UpdateAllViews()
			a.App.SetFocus(a.CurrentTableView)
		}
	}
}
