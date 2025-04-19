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

	for _, entity := range model.SACCTMGR_TABLE_ENTITIES {
		a.SacctMgrEntitySelector.AddOption(
			entity,
			a.applySacctMgrEntitySelector(entity),
		)
	}
	a.SacctMgrEntitySelector.SetCurrentOption(0)
}

func (a *App) applySacctMgrEntitySelector(entity string) func() {
	return func() {
		config.SacctMgrCurrentEntity = entity
		a.SacctMgrProvider.Fetch()
		if a.FirstRenderComplete {
			a.SacctMgrView.SetTitleHeader(entity)
			a.SacctMgrView.Render()
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
		}
	}
}
