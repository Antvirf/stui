package view

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupSacctMgrEntitySelector() {
	a.SacctMgrEntitySelector = tview.NewDropDown().
		SetLabel("(e) Entity: ").
		SetLabelStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite)).
		SetListStyles(
			tcell.StyleDefault,
			tcell.StyleDefault.Background(selectionColor),
		).
		SetFieldWidth(20).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetTextOptions("  ", "  ", "", "", "")

	a.SacctMgrEntitySelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.App.SetFocus(a.CurrentTableView)
			return nil
		}
		return event
	})

	for _, entity := range model.SACCTMGR_TABLE_ENTITIES {
		a.SacctMgrEntitySelector.AddOption(
			entity,
			a.applySacctMgrEntitySelector(entity),
		)
	}
	a.SacctMgrEntitySelector.SetCurrentOption(0)
}

func (a *App) DisableSacctMgrEntitySelector(disabled bool) {
	if config.SacctEnabled {
		a.SacctMgrEntitySelector.SetDisabled(disabled)
	}
}

func (a *App) applySacctMgrEntitySelector(entity string) func() {
	return func() {
		config.SacctMgrCurrentEntity = entity
		a.SacctMgrProvider.Fetch()
		if a.FirstRenderComplete {
			a.UpdateAllViews()
			a.App.SetFocus(a.CurrentTableView)
		}
	}
}
