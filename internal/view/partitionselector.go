package view

import (
	"fmt"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	ALL_PARTITIONS_OPTION = "(all)"
)

func (a *App) SetupPartitionSelector() {
	// Partition selector
	a.PartitionSelector = tview.NewDropDown().
		SetLabel("Partition: ").
		SetLabelStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite)).
		SetListStyles(
			tcell.StyleDefault,
			tcell.StyleDefault.Background(selectionColor),
		).
		SetFieldWidth(20).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetTextOptions("  ", "  ", "", "", "")

	a.PartitionSelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.App.SetFocus(a.CurrentTableView)
			return nil
		}
		return event
	})
}

func (a *App) setupPartitionSelectorOptions() {
	for index, partition := range a.PartitionsData.Rows {
		if index == 0 {
			a.PartitionSelector.AddOption(
				ALL_PARTITIONS_OPTION,
				a.applyPartitionSelector(ALL_PARTITIONS_OPTION),
			)
		}

		partitionName := partition[0]
		a.PartitionSelector.AddOption(
			partitionName,
			a.applyPartitionSelector(partitionName),
		)
	}

	// Set selected option at start
	if config.PartitionFilter == "" {
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
			config.PartitionFilter = ""
		}
	}
}

func (a *App) applyPartitionSelector(partition string) func() {
	return func() {
		if partition == ALL_PARTITIONS_OPTION {
			config.PartitionFilter = ""
		} else {
			config.PartitionFilter = partition
		}
		if !a.PartitionSelectorFirstUpdate {
			// On the first update, we do not need to refetch data as
			// fetcher functions took care of that already
			a.UpdateAllViews()
		}
		a.PartitionSelectorFirstUpdate = false
		a.App.SetFocus(a.CurrentTableView)
	}
}
