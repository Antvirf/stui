package view

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupSacctMgrEntitySelector(storedSelectorValue string) {
	a.SacctMgrEntitySelector = tview.NewDropDown().
		SetLabel(PadSelectorTitle("(e) Entity:")).
		SetLabelStyle(tcell.StyleDefault.Foreground(dropdownForegroundColor)).
		SetListStyles(
			tcell.StyleDefault,
			tcell.StyleDefault.Background(selectionColor),
		).
		SetFieldWidth(20).
		SetFieldBackgroundColor(dropdownBackgroundColor).
		SetTextOptions("  ", "  ", "", "", "")

	a.SacctMgrEntitySelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
			return nil
		}
		return event
	})

	for index, entity := range model.SACCTMGR_TABLE_ENTITIES {
		a.SacctMgrEntitySelector.AddOption(
			entity,
			a.applySacctMgrEntitySelector(entity),
		)

		// By default, show 0 index
		if index == 0 {
			// This line leads to a call being made to Sacctmgr during setup time
			a.SacctMgrEntitySelector.SetCurrentOption(0)
		}

		// If storedSelectorValue matches anything, use that.
		if entity == storedSelectorValue {
			a.SacctMgrEntitySelector.SetCurrentOption(index)
		}
	}
}

func (a *App) applySacctMgrEntitySelector(entity string) func() {
	return func() {
		config.SacctMgrCurrentEntity = entity
		a.SacctMgrProvider.Fetch()
		if a.FirstRenderComplete {
			a.SacctMgrView.SetTitleHeader(entity)
			a.setupSortSelectorOptions(a.SacctMgrProvider, a.SacctMgrView.sortColumn)
			a.SacctMgrView.Render()
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
		}
	}
}
