package view

import (
	"fmt"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) SetupPartitionSelector() {
	// Partition selector
	a.PartitionSelector = tview.NewDropDown().
		SetLabel("(p) Partition: ").
		SetLabelStyle(tcell.StyleDefault.Foreground(dropdownForegroundColor)).
		SetListStyles(
			tcell.StyleDefault,
			tcell.StyleDefault.Background(selectionColor),
		).
		SetFieldWidth(20).
		SetFieldBackgroundColor(dropdownBackgroundColor).
		SetTextOptions("  ", "  ", "", "", "")

	a.PartitionSelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			_, frontPage := a.Pages.GetFrontPage()
			a.App.SetFocus(frontPage)
			return nil
		}
		return event
	})
}

func (a *App) setupPartitionSelectorOptions() {
	for index, partition := range a.PartitionsData.Rows {
		if index == 0 {
			a.PartitionSelector.AddOption(
				config.ALL_CATEGORIES_OPTION,
				a.applyPartitionSelector(config.ALL_CATEGORIES_OPTION),
			)
		}

		partitionName := partition[0]
		a.PartitionSelector.AddOption(
			partitionName,
			a.applyPartitionSelector(partitionName),
		)
	}

	// Set selected option at start
	if config.PartitionFilter == config.ALL_CATEGORIES_OPTION {
		a.PartitionSelector.SetCurrentOption(0)
	} else {
		found := false
		for index, partition := range a.PartitionsData.Rows {
			if partition[0] == config.PartitionFilter {
				a.PartitionSelector.SetCurrentOption(index + 1)
				found = true
				break
			}
		}
		if !found {
			a.ShowNotification(
				fmt.Sprintf("[red]Requested partition '%s' does not exist, using no filter instead[white]", config.PartitionFilter),
				2*time.Second,
			)
			a.PartitionSelector.SetCurrentOption(0)
			config.PartitionFilter = config.ALL_CATEGORIES_OPTION
		}
	}
}

func (a *App) applyPartitionSelector(partition string) func() {
	return func() {
		if partition == config.ALL_CATEGORIES_OPTION {
			config.PartitionFilter = config.ALL_CATEGORIES_OPTION
		} else {
			config.PartitionFilter = partition
		}
		a.RenderCurrentView()
		_, frontPage := a.Pages.GetFrontPage()
		a.App.SetFocus(frontPage)
	}
}
