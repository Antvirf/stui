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
		SetLabel(PadSelectorTitle("(p) Partition:")).
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

func (a *App) setupPartitionSelectorOptions(storedSelectorValue string) {
	if storedSelectorValue == "" { // If empty, treat it equal to '(all)'
		storedSelectorValue = config.ALL_CATEGORIES_OPTION
	}
	for index, partition := range a.PartitionsData.Rows {
		if index == 0 {
			a.PartitionSelector.AddOption(
				config.ALL_CATEGORIES_OPTION,
				a.applyPartitionSelector(config.ALL_CATEGORIES_OPTION),
			)
		}

		partitionName := partition[0].Display()
		a.PartitionSelector.AddOption(
			partitionName,
			a.applyPartitionSelector(partitionName),
		)
	}

	if config.PartitionFilter == config.ALL_CATEGORIES_OPTION && storedSelectorValue == config.ALL_CATEGORIES_OPTION {
		a.PartitionSelector.SetCurrentOption(0)
		return
	}

	// Loop through partitions options, if user has given a commmand-line arg,
	// use that first, otherwise revert to stored state if available.
	storedSelectorIndex := 0
	requestedSelectorIndex := 0
	for index, partition := range a.PartitionsData.Rows {
		if partition[0].Display() == config.PartitionFilter {
			requestedSelectorIndex = index + 1
			break // no need to go further
		}
		if partition[0].Display() == storedSelectorValue {
			storedSelectorIndex = index + 1
		}
	}

	// If requested available, use it.
	// Alternatively, if stored was available, use it
	// Otherwise, set to (all).
	if requestedSelectorIndex != 0 {
		a.PartitionSelector.SetCurrentOption(requestedSelectorIndex)
	} else if storedSelectorIndex != 0 {
		a.PartitionSelector.SetCurrentOption(storedSelectorIndex)
	} else {
		a.ShowNotification(
			fmt.Sprintf("[red]Requested partition '%s' does not exist, no partition filter applied[white]", config.PartitionFilter),
			2*time.Second,
		)
		a.PartitionSelector.SetCurrentOption(0)
		config.PartitionFilter = config.ALL_CATEGORIES_OPTION
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
