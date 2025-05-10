package view

import (
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupSortSelector() {
	a.SortSelector = tview.NewDropDown().
		SetLabel(PadSelectorTitle("(o) Sort by:")).
		SetLabelStyle(tcell.StyleDefault.Foreground(dropdownForegroundColor)).
		SetListStyles(
			tcell.StyleDefault,
			tcell.StyleDefault.Background(selectionColor),
		).
		SetFieldWidth(20).
		SetFieldBackgroundColor(dropdownBackgroundColor).
		SetTextOptions("  ", "  ", "", "", "")

	a.SortSelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
			return nil
		}
		return event
	})
}

func (a *App) setupSortSelectorOptions(provider model.DataProvider[*model.TableData], selectedColumn int) {
	a.SortSelector.SetCurrentOption(selectedColumn)
	a.SortSelector.SetOptions([]string{}, nil)
	columns := provider.Data().Headers
	for index, column := range *columns {
		if index == 0 {
			a.SortSelector.AddOption(
				config.NO_SORT_OPTION,
				a.applySortSelector(-1),
			)
		}
		a.SortSelector.AddOption(
			column.Name,
			a.applySortSelector(index),
		)
	}

	if selectedColumn == -1 {
		a.SortSelector.SetCurrentOption(0)
	} else {
		a.SortSelector.SetCurrentOption(selectedColumn + 1)
	}
}

func (a *App) applySortSelector(column int) func() {
	return func() {
		view := a.GetCurrentStuiView()
		if view == nil {
			return
		}

		view.sortColumn = column

		// Reverse sort direction if the same column is selected again
		switch view.sortDirection {
		case SORT_NONE:
			view.sortDirection = SORT_ASC
		case SORT_ASC:
			view.sortDirection = SORT_DESC
		case SORT_DESC:
			view.sortDirection = SORT_ASC
		}

		a.RenderCurrentView()
		_, frontPage := a.Pages.GetFrontPage()
		a.App.SetFocus(frontPage)
	}
}
